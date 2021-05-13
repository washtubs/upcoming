package upcoming

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

type UpcomingClient struct {
	prefix string
	client *redis.Client
}

func DefaultClient() *UpcomingClient {
	return &UpcomingClient{
		prefix: "upcoming",
		client: redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "", // no password set
			DB:       0,  // use default DB
		}),
	}
}

type ListOpts struct {
	Within  time.Duration
	Sources []string
}

type Upcoming struct {
	Source   string    `json:"source"`
	SourceId string    `json:"sourceId"`
	Title    string    `json:"title"`
	When     time.Time `json:"when"`
}

func (u Upcoming) HumanizeDuration() string {
	return HumanizeDuration(time.Until(u.When))
}

func (u *UpcomingClient) encodeUpcoming(c Upcoming) []byte {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf) // TODO: Encoder should be reused somehow. It's difficult without sacking thread safety though
	err := enc.Encode(c)
	if err != nil {
		panic(errors.WithMessage(err, "Failed to encode context"))
	}

	return buf.Bytes()
}

func (u *UpcomingClient) decodeUpcoming(iface interface{}) Upcoming {
	str, ok := iface.(string)
	if !ok {
		panic("Failed to convert value to string")
	}
	return u.decodeUpcomingBuf([]byte(str))
}

func (r *UpcomingClient) decodeUpcomingBuf(buf []byte) Upcoming {
	dec := json.NewDecoder(bytes.NewBuffer(buf))
	var upcoming Upcoming
	err := dec.Decode(&upcoming)
	if err != nil {
		panic(errors.WithMessage(err, "Failed to decode context"))
	}

	return upcoming
}

func (u *UpcomingClient) list(path string) ([]Upcoming, error) {
	keys, err := u.client.Keys(context.Background(), path).Result()
	if err != nil {
		if err == redis.Nil {
			return []Upcoming{}, nil
		}
		return []Upcoming{}, errors.Wrapf(err, "Failed to obtain keys")
	}
	if len(keys) == 0 {
		return []Upcoming{}, nil
	}

	res, err := u.client.MGet(context.Background(), keys...).Result()
	if err != nil {
		return []Upcoming{}, errors.Wrapf(err, "Failed to obtain objects at keys: %+v", keys)
	}

	upcomings := make([]Upcoming, 0, len(keys))
	for _, v := range res {
		ctx := u.decodeUpcoming(v)
		upcomings = append(upcomings, ctx)
	}
	return upcomings, nil
}

type Upcomings []Upcoming

func (s Upcomings) Len() int {
	return len(s)
}
func (s Upcomings) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type upcomingByDuration struct{ Upcomings }

func (s *upcomingByDuration) Less(i, j int) bool {
	return s.Upcomings[i].When.Before(s.Upcomings[j].When)
}

func (u *UpcomingClient) List(opts ListOpts) (list []Upcoming, err error) {
	if opts.Sources != nil && len(opts.Sources) > 0 {
		all := make([]Upcoming, 0)
		for _, s := range opts.Sources {
			l, err := u.list(path.Join(u.prefix, s, "*"))
			if err != nil {
				return []Upcoming{}, errors.Wrapf(err, "Failed to obtain list for %s", s)
			}
			all = append(all, l...)
		}
		list, err = all, nil
	} else {
		list, err = u.list(path.Join(u.prefix, "*"))
	}

	sort.Sort(sort.Reverse(&upcomingByDuration{list}))
	return list, err
}

func (u *UpcomingClient) RemoveAll(source string) (int64, error) {

	keys, err := u.client.Keys(context.Background(), path.Join(u.prefix, source, "*")).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, errors.Wrapf(err, "Failed to obtain keys")
	}
	if len(keys) == 0 {
		return 0, nil
	}

	deleted := int64(0)
	for _, key := range keys {
		val, err := u.client.Del(context.Background(), key).Result()
		if err != nil {
			if err == redis.Nil {
				// IDK if this is possible
				continue
			}
			return deleted, errors.Wrapf(err, "Failed to delete existing upcoming %s", key)
		}
		deleted = deleted + val
	}
	return deleted, nil
}

func (u *UpcomingClient) Remove(source, sourceId string) (bool, error) {
	key := path.Join(u.prefix, source, sourceId)
	val, err := u.client.Del(context.Background(), key).Result()
	if err != nil {
		if err == redis.Nil {
			// IDK if this is possible
			return false, nil
		}
		return false, errors.Wrapf(err, "Failed to delete existing upcoming %s", key)
	}
	return val == 1, nil
}

func (u *UpcomingClient) Put(upcoming Upcoming) error {
	if upcoming.When.Before(time.Now()) {
		// Nothing to do. It's not upcoming anymore
		return nil
	}
	key := path.Join(u.prefix, upcoming.Source, upcoming.SourceId)

	err := u.client.Set(context.Background(), key, u.encodeUpcoming(upcoming), time.Until(upcoming.When)).Err()
	if err != nil {
		return errors.Wrapf(err, "Failed to set upcoming at %s", key)
	}
	return nil

}

func Format(upcoming Upcoming) string {
	return fmt.Sprintf("%s\t%s\t%s\t%s", HumanizeDuration(time.Until(upcoming.When)), upcoming.Title, upcoming.Source, upcoming.SourceId)
}
