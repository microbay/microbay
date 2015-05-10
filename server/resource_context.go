package server

import (
	"github.com/SHMEDIALIMITED/apigo/model"
	log "github.com/Sirupsen/logrus"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gocraft/web"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

const (
	REDIS string = "redis"
)

type ResourceContext struct {
	*Context
	Resource *model.Resource
}

func (c *ResourceContext) ResourceConfigMiddleware(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	c.Resource = c.Config.FindResourceByPath(req.PathParams["resource"])
	if c.Resource == nil {
		c.RenderError(rw, "Access Forbidden", http.StatusForbidden)
	} else {
		next(rw, req)
	}
}

func (c *ResourceContext) RedisToJWTAuthMiddleware(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	if c.Resource.Auth == REDIS {
		c.RenderError(rw, "Invalid token", http.StatusUnauthorized)
	} else {
		next(rw, req)
	}
}

func (c *ResourceContext) BalancedProxy(rw web.ResponseWriter, req *web.Request) {

	backend := c.Resource.Backends.Choose()
	if backend == nil {
		log.Error("no backend for client %s", req.RemoteAddr)
	}

	serverUrl, err := url.Parse(backend.String())
	if err != nil {
		log.Fatal("URL failed to parse")
	}

	// initialize our reverse proxy
	reverseProxy := httputil.NewSingleHostReverseProxy(serverUrl)

	combinedHeaders := headerCombiner(reverseProxy, c.Config.Key)

	req.URL.Path = ""
	combinedHeaders.ServeHTTP(rw, req.Request)
}

// Append additional query params to the original URL query.
func headerCombiner(handler http.Handler, key []byte) http.Handler {

	c := map[string]interface{}{"premise_id": -2287340764, "exp": float64(time.Now().Unix() + 100)}

	token := jwt.New(jwt.SigningMethodRS256)
	token.Claims = c
	tokenString, e := token.SignedString(key)

	if e != nil {
		panic(e.Error())
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Authorization", tokenString)
		handler.ServeHTTP(w, r)
	})
}
