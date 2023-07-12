package render

import (
	"encoding/gob"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/Ed-cred/bookings/internal/config"
	"github.com/Ed-cred/bookings/internal/models"
	"github.com/alexedwards/scs/v2"
)

var (
	session *scs.SessionManager
	testApp config.AppConfig
)

func TestMain(m *testing.M) {

	gob.Register(models.Reservation{})

	// change to true when in produciton
	testApp.InProd = false
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = testApp.InProd

	testApp.Session = session
	app = &testApp

	os.Exit(m.Run())
}

type myWriter struct {

}

func (w *myWriter) Header() http.Header {
	var head http.Header
	return head
}

func (w *myWriter) Write(p []byte) (int, error) {
	length := len(p)
	return length, nil
}

func (w *myWriter) WriteHeader(i int) {

}
