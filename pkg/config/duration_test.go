package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDurationJSON(t *testing.T) {
	t.Run("Unmarshal", func(t *testing.T) {
		assert := assert.New(t)

		buf := []byte(`"32h15m31s"`)
		var res Duration
		err := (&res).UnmarshalJSON(buf)

		d := Duration(32*time.Hour + 15*time.Minute + 31*time.Second)

		assert.NoError(err, "Should convert from json without error")
		assert.Equal(d, res, "Should be the correct duration")
	})
	t.Run("Marshal", func(t *testing.T) {
		assert := assert.New(t)

		in := Duration(time.Hour + time.Minute*5 + time.Second*21)

		buf, err := in.MarshalJSON()

		assert.NoError(err, "Should convert to json without error")
		assert.Equal(`"1h5m21s"`, string(buf), "Should convert to the correct json string")
	})
}
