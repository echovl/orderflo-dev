package http

import (
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// initRoutes registers server routes and handlers
func (s *Server) initRoutes() {
	s.App.Use(cors.New())
	s.App.Use(s.loggerHandler)

	api := s.App.Group("/api/v1")
	root := s.App.Group("/")

	root.Get("/:id", s.handleRenderDesign)

	api.Get("/health", s.handleCheckHealth)

	api.Post("/auth/signup", s.handleSignUp)
	api.Post("/auth/signin", s.handleSignIn)
	api.Post("/auth/signout", s.requireUserSession, s.handleSignOut)
	api.Get("/auth/me", s.requireUserSession, s.handleCurrentUser)
	api.Put("/auth/profile", s.requireUserSession, s.handleUpdateUserProfile)
	api.Get("/auth/signin/github", s.handleGithubSignIn)
	api.Get("/auth/signin/google", s.handleGoogleSignIn)
	api.Get("/auth/callback/github", s.handleGithubCallback)
	api.Get("/auth/callback/google", s.handleGoogleCallback)
	api.Get("/auth/csrf", s.requireUserSession, s.handleGetCSRFToken)

	api.Get("/companies", s.requireUserSession, s.handleListCompanies)
	api.Get("/companies/:id", s.requireUserSession, s.handleGetCompany)
	api.Post("/companies", s.requireUserSession, s.handleCreateCompany)
	api.Put("/companies/:id", s.requireUserSession, s.handleUpdateCompany)
	api.Delete("/companies/:id", s.requireUserSession, s.handleDeleteCompany)

	api.Get("/customers", s.requireUserSession, s.handleListCustomers)
	api.Get("/customers/:id", s.requireUserSession, s.handleGetCustomer)
	api.Post("/customers", s.requireUserSession, s.handleCreateCustomer)
	api.Put("/customers/:id", s.requireUserSession, s.handleUpdateCustomer)
	api.Delete("/customers/:id", s.requireUserSession, s.handleDeleteCustomer)

	api.Get("/frames", s.requireUserSession, s.handleListFrames)
	api.Get("/frames/:id", s.requireUserSession, s.handleGetFrame)
	api.Post("/frames", s.requireUserSession, s.requireAdmin, s.handleCreateFrame)
	api.Put("/frames/:id", s.requireUserSession, s.requireAdmin, s.handleUpdateFrame)
	api.Delete("/frames/:id", s.requireUserSession, s.requireAdmin, s.handleDeleteFrame)

	api.Get("/templates", s.requireUserSession, s.handleListTemplate)
	api.Get("/templates/:id", s.requireUserSession, s.handleGetTemplate)
	api.Post("/templates", s.requireUserSession, s.requireAdmin, s.handleCreateTemplate)
	api.Put("/templates/:id", s.requireUserSession, s.requireAdmin, s.handleUpdateTemplate)
	api.Delete("/templates/:id", s.requireUserSession, s.requireAdmin, s.handleDeleteTemplate)

	api.Get("/render/:id", s.requireCompanySession, s.handleRenderDesign)

	api.Get("/projects", s.requireUserSession, s.handleListProject)
	api.Get("/projects/:id", s.requireUserSession, s.handleGetProject)
	api.Post("/projects", s.requireUserSession, s.handleCreateProject)
	api.Put("/projects/:id", s.requireUserSession, s.handleUpdateProject)
	api.Delete("/projects/:id", s.requireUserSession, s.handleDeleteProject)

	api.Get("/components", s.requireUserSession, s.handleListComponent)
	api.Get("/components/:id", s.requireUserSession, s.handleGetComponent)
	api.Post("/components", s.requireUserSession, s.handleCreateComponent)
	api.Put("/components/:id", s.requireUserSession, s.handleUpdateComponent)
	api.Delete("/components/:id", s.requireUserSession, s.handleDeleteComponent)

	api.Post("/uploads", s.requireUserSession, s.handleCreateSignedURL)
	api.Put("/uploads", s.requireUserSession, s.handleCreateUpload)
	api.Get("/uploads", s.requireUserSession, s.handleListUpload)
	api.Delete("/uploads/:id", s.requireUserSession, s.handleDeleteUpload)

	api.Get("/resources/pixabay/images", s.handleFetchPixabayImages)
	api.Get("/resources/pixabay/videos", s.handleFetchPixabayVideos)
	api.Get("/resources/pexels/images", s.handleFetchPexelsImages)
	api.Get("/resources/pexels/videos", s.handleFetchPexelsVideos)

	api.Get("/fonts", s.requireUserSession, s.handleListFonts)
	api.Get("/fonts/:id", s.requireUserSession, s.handleGetFont)
	api.Post("/fonts", s.requireUserSession, s.handleCreateFont)
	api.Put("/fonts/:id", s.requireUserSession, s.handleUpdateFont)
	api.Delete("/fonts/:id", s.requireUserSession, s.handleDeleteFont)
	api.Post("/fonts/enable", s.requireUserSession, s.handleEnableFonts)
	api.Post("/fonts/disable", s.requireUserSession, s.handleDisableFonts)

	api.Post("/subscriptions/products", s.requireUserSession, s.requireAdmin, s.handleCreateProduct)
	api.Get("/subscriptions/products", s.requireUserSession, s.requireAdmin, s.handleListProducts)
	api.Get("/subscriptions/plans", s.handleListPlan)
	api.Post("/subscriptions/plans", s.requireUserSession, s.requireAdmin, s.handleCreatePlan)
	api.Post("/subscriptions/plans/:id/subscribe", s.requireUserSession, s.handleSubscribeUser)
}
