package api

import (
	"embed"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Server represents the HTTP/WebSocket server
type Server struct {
	router   *gin.Engine
	port     int
	staticFS embed.FS
	devMode  bool
}

// NewServer creates a new API server
func NewServer(port int, staticFS embed.FS, devMode bool) *Server {
	if port == 0 {
		port = 8080
	}

	if devMode {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.SetTrustedProxies(nil)
	router.Use(gin.Recovery())

	return &Server{
		router:   router,
		port:     port,
		staticFS: staticFS,
		devMode:  devMode,
	}
}

// Router returns the gin router for route registration
func (s *Server) Router() *gin.Engine {
	return s.router
}

// Start begins listening for connections
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	return s.router.Run(addr)
}

// ServeStatic serves the embedded Vue frontend (skipped in dev mode)
func (s *Server) ServeStatic() {
	if s.devMode {
		return
	}

	// Serve static files from embedded FS
	httpFS := http.FS(s.staticFS)

	// Serve index.html for root path
	s.router.GET("/", func(c *gin.Context) {
		data, err := s.staticFS.ReadFile("static/index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load frontend")
			return
		}
		c.Data(http.StatusOK, "text/html", data)
	})

	// Serve assets
	s.router.GET("/assets/*filepath", func(c *gin.Context) {
		c.FileFromFS("static/assets/"+c.Param("filepath"), httpFS)
	})
}
