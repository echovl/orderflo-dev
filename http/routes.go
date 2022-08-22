package http

import (
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// initRoutes registers server routes and handlers
func (s *Server) initRoutes() {
	s.App.Use(cors.New())
	s.App.Use(s.loggerHandler)

	web := s.App.Group("/web")
	editor := s.App.Group("/editor")
	root := s.App.Group("/")

	root.Get("/:id", s.handleRenderDesign)
	root.Get("/health", s.handleCheckHealth)

	editor.Post("/auth/signup", s.handleCustomerSignUp)
	editor.Post("/auth/signin", s.handleCustomerSignIn)

	editor.Get("/customers/me", s.requireCustomerSession, s.handleCurrentCustomer)
	editor.Put("/customers/:id", s.requireCustomerSession, s.handleUpdateCustomer)

	editor.Get("/projects", s.requireCustomerSession, s.handleListProject)
	editor.Get("/projects/:id", s.requireCustomerSession, s.handleGetProject)
	editor.Post("/projects", s.requireCustomerSession, s.handleCreateProject)
	editor.Put("/projects/:id", s.requireCustomerSession, s.handleUpdateProject)
	editor.Delete("/projects/:id", s.requireCustomerSession, s.handleDeleteProject)

	editor.Get("/fonts", s.requireCustomerSession, s.handleListFonts)
	editor.Get("/fonts/:id", s.requireCustomerSession, s.handleGetFont)
	editor.Post("/fonts", s.requireCustomerSession, s.handleCreateFont)
	editor.Put("/fonts/:id", s.requireCustomerSession, s.handleUpdateFont)
	editor.Delete("/fonts/:id", s.requireCustomerSession, s.handleDeleteFont)
	editor.Post("/fonts/enable", s.requireCustomerSession, s.handleEnableFonts)
	editor.Post("/fonts/disable", s.requireCustomerSession, s.handleDisableFonts)

	editor.Post("/uploads", s.requireCustomerSession, s.handleCreateSignedURL)
	editor.Put("/uploads", s.requireCustomerSession, s.handleCreateUpload)
	editor.Get("/uploads", s.requireCustomerSession, s.handleListUpload)
	editor.Delete("/uploads/:id", s.requireCustomerSession, s.handleDeleteUpload)

	editor.Get("/frames", s.requireCustomerSession, s.handleListFrames)
	editor.Get("/frames/:id", s.requireCustomerSession, s.handleGetFrame)
	editor.Post("/frames", s.requireCustomerSession, s.handleCreateFrame)
	editor.Put("/frames/:id", s.requireCustomerSession, s.handleUpdateFrame)
	editor.Delete("/frames/:id", s.requireCustomerSession, s.handleDeleteFrame)

	editor.Get("/resources/pixabay/images", s.requireCustomerSession, s.handleFetchPixabayImages)
	editor.Get("/resources/pixabay/videos", s.requireCustomerSession, s.handleFetchPixabayVideos)
	editor.Get("/resources/pexels/images", s.requireCustomerSession, s.handleFetchPexelsImages)
	editor.Get("/resources/pexels/videos", s.requireCustomerSession, s.handleFetchPexelsVideos)

	web.Post("/auth/signup", s.handleUserSignUp)
	web.Post("/auth/signin", s.handleUserSignIn)
	web.Post("/auth/signout", s.requireUserSession, s.handleSignOut)
	web.Get("/auth/me", s.requireUserSession, s.handleCurrentUser)
	web.Put("/auth/profile", s.requireUserSession, s.handleUpdateUserProfile)
	web.Get("/auth/signin/github", s.handleGithubSignIn)
	web.Get("/auth/signin/google", s.handleGoogleSignIn)
	web.Get("/auth/callback/github", s.handleGithubCallback)
	web.Get("/auth/callback/google", s.handleGoogleCallback)
	web.Get("/auth/csrf", s.requireUserSession, s.handleGetCSRFToken)

	web.Get("/companies/:id", s.requireUserSession, s.handleGetCompany)
	web.Put("/companies/:id", s.requireUserSession, s.handleUpdateCompany)

	web.Get("/customers", s.requireUserSession, s.handleListCustomers)
	web.Get("/customers/:id", s.requireUserSession, s.handleGetCustomer)
	web.Put("/customers/:id", s.requireUserSession, s.handleUpdateCustomer)
	web.Delete("/customers/:id", s.requireUserSession, s.handleDeleteCustomer)

	web.Get("/frames", s.requireUserSession, s.handleListFrames)
	web.Get("/frames/:id", s.requireUserSession, s.handleGetFrame)
	web.Post("/frames", s.requireUserSession, s.handleCreateFrame)
	web.Put("/frames/:id", s.requireUserSession, s.handleUpdateFrame)
	web.Delete("/frames/:id", s.requireUserSession, s.handleDeleteFrame)

	web.Get("/templates", s.requireUserSession, s.handleListTemplate)
	web.Get("/templates/:id", s.requireUserSession, s.handleGetTemplate)
	web.Post("/templates", s.requireUserSession, s.handleCreateTemplate)
	web.Put("/templates/:id", s.requireUserSession, s.handleUpdateTemplate)
	web.Delete("/templates/:id", s.requireUserSession, s.handleDeleteTemplate)

	web.Get("/render/:id", s.handleRenderDesign)

	web.Get("/projects", s.requireUserSession, s.handleListProject)
	web.Get("/projects/:id", s.requireUserSession, s.handleGetProject)
	web.Post("/projects", s.requireUserSession, s.handleCreateProject)
	web.Put("/projects/:id", s.requireUserSession, s.handleUpdateProject)
	web.Delete("/projects/:id", s.requireUserSession, s.handleDeleteProject)

	web.Get("/components", s.requireUserSession, s.handleListComponent)
	web.Get("/components/:id", s.requireUserSession, s.handleGetComponent)
	web.Post("/components", s.requireUserSession, s.handleCreateComponent)
	web.Put("/components/:id", s.requireUserSession, s.handleUpdateComponent)
	web.Delete("/components/:id", s.requireUserSession, s.handleDeleteComponent)

	web.Post("/uploads", s.requireUserSession, s.handleCreateSignedURL)
	web.Put("/uploads", s.requireUserSession, s.handleCreateUpload)
	web.Get("/uploads", s.requireUserSession, s.handleListUpload)
	web.Delete("/uploads/:id", s.requireUserSession, s.handleDeleteUpload)

	web.Get("/resources/pixabay/images", s.handleFetchPixabayImages)
	web.Get("/resources/pixabay/videos", s.handleFetchPixabayVideos)
	web.Get("/resources/pexels/images", s.handleFetchPexelsImages)
	web.Get("/resources/pexels/videos", s.handleFetchPexelsVideos)

	web.Get("/fonts", s.requireUserSession, s.handleListFonts)
	web.Get("/fonts/:id", s.requireUserSession, s.handleGetFont)
	web.Post("/fonts", s.requireUserSession, s.handleCreateFont)
	web.Put("/fonts/:id", s.requireUserSession, s.handleUpdateFont)
	web.Delete("/fonts/:id", s.requireUserSession, s.handleDeleteFont)
	web.Post("/fonts/enable", s.requireUserSession, s.handleEnableFonts)
	web.Post("/fonts/disable", s.requireUserSession, s.handleDisableFonts)
}
