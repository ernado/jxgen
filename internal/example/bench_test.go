package example

import (
	stdlib "encoding/json"
	"io"
	"testing"

	"github.com/go-faster/jx"
	"github.com/goccy/go-json"
	"github.com/mailru/easyjson/jwriter"
)

func BenchmarkEncoding(b *testing.B) {
	s := &Struct{
		Name:  "kek",
		Value: 42,
	}
	b.Run("jxgen", func(b *testing.B) {
		b.ReportAllocs()

		w := jx.GetWriter()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w.Reset()
			s.WriteJSON(w)
		}
	})
	b.Run("goccy", func(b *testing.B) {
		b.ReportAllocs()
		e := json.NewEncoder(io.Discard)
		e.SetEscapeHTML(false)

		for i := 0; i < b.N; i++ {
			if err := e.Encode(s); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("stdlib", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			if _, err := stdlib.Marshal(s); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("easyjson", func(b *testing.B) {
		b.ReportAllocs()

		w := &jwriter.Writer{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w.Buffer.Buf = w.Buffer.Buf[:0]
			s.MarshalEasyJSON(w)
		}
	})
}
