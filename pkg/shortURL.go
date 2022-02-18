package pkg

import (
	crand "crypto/rand"
	"encoding/binary"
	"github.com/rs/zerolog"
	rand "math/rand"
	"os"
)

type cryptoSource struct{}

func (s cryptoSource) Seed(seed int64) {}

func (s cryptoSource) Int63() int64 {
	return int64(s.Uint64() & ^uint64(1<<63))
}

func (s cryptoSource) Uint64() (v uint64) {
	logger := zerolog.New(os.Stdout).With().
		Timestamp().
		Logger()

	logger = logger.Output(zerolog.NewConsoleWriter())

	err := binary.Read(crand.Reader, binary.BigEndian, &v)
	if err != nil {
		logger.Fatal().Err(err).Msg("cannot creat short url")
	}
	return v
}

func GeneratorShortURL() string {
	DIGITS := "0123456789"
	UPPERCASE := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	LOWERCASE := "abcdefghijklmnopqrstuvwxyz"
	UNDERSCORE := "_"
	ALL := UPPERCASE + LOWERCASE + DIGITS + UNDERSCORE
	length := 10

	var src cryptoSource
	rnd := rand.New(src)

	buf := make([]byte, length)
	buf[0] = DIGITS[rnd.Intn(len(DIGITS))]
	buf[1] = UPPERCASE[rnd.Intn(len(UPPERCASE))]
	buf[2] = LOWERCASE[rnd.Intn(len(LOWERCASE))]
	buf[3] = UNDERSCORE[0]

	for i := 4; i < length; i++ {
		buf[i] = ALL[rnd.Intn(len(ALL))]
	}
	rnd.Shuffle(len(buf), func(i, j int) {
		buf[i], buf[j] = buf[j], buf[i]
	})
	str := string(buf)
	return str
}
