package http

import (
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// initRoutes registers server routes and handlers
func (s *Server) initRoutes() {
	s.App.Use(cors.New())
	s.App.Use(s.loggerHandler)

	web := s.App.Group("/web")
	api := s.App.Group("/api")
	root := s.App.Group("/")

	root.Get("/:id", s.handleRenderDesign)
	root.Get("/health", s.handleCheckHealth)

	api.Get("/customers", s.requireCompanySession, s.handleListCustomers)
	api.Get("/customers/:id", s.requireCompanySession, s.handleGetCustomer)
	api.Post("/customers", s.requireCompanySession, s.handleCreateCustomer)
	api.Put("/customers/:id", s.requireCompanySession, s.handleUpdateCustomer)
	api.Delete("/customers/:id", s.requireCompanySession, s.handleDeleteCustomer)

	api.Get("/projects", s.requireCompanySession, s.handleListProject)
	api.Get("/projects/:id", s.requireCompanySession, s.handleGetProject)
	api.Post("/projects", s.requireCompanySession, s.handleCreateProject)
	api.Put("/projects/:id", s.requireCompanySession, s.handleUpdateProject)
	api.Delete("/projects/:id", s.requireCompanySession, s.handleDeleteProject)

	api.Get("/fonts", s.requireCompanySession, s.handleListFonts)
	api.Get("/fonts/:id", s.requireCompanySession, s.handleGetFont)
	api.Post("/fonts", s.requireCompanySession, s.handleCreateFont)
	api.Put("/fonts/:id", s.requireCompanySession, s.handleUpdateFont)
	api.Delete("/fonts/:id", s.requireCompanySession, s.handleDeleteFont)
	api.Post("/fonts/enable", s.requireCompanySession, s.handleEnableFonts)
	api.Post("/fonts/disable", s.requireCompanySession, s.handleDisableFonts)

	api.Get("/frames", s.requireUserSession, s.handleListFrames)
	api.Get("/frames/:id", s.requireUserSession, s.handleGetFrame)

	web.Post("/auth/signup", s.handleSignUp)
	web.Post("/auth/signin", s.handleSignIn)
	web.Post("/auth/signout", s.requireUserSession, s.handleSignOut)
	web.Get("/auth/me", s.requireUserSession, s.handleCurrentUser)
	web.Put("/auth/profile", s.requireUserSession, s.handleUpdateUserProfile)
	web.Get("/auth/signin/github", s.handleGithubSignIn)
	web.Get("/auth/signin/google", s.handleGoogleSignIn)
	web.Get("/auth/callback/github", s.handleGithubCallback)
	web.Get("/auth/callback/google", s.handleGoogleCallback)
	web.Get("/auth/csrf", s.requireUserSession, s.handleGetCSRFToken)

	web.Get("/companies", s.requireUserSession, s.handleListCompanies)
	web.Get("/companies/:id", s.requireUserSession, s.handleGetCompany)
	web.Post("/companies", s.requireUserSession, s.handleCreateCompany)
	web.Put("/companies/:id", s.requireUserSession, s.handleUpdateCompany)
	web.Delete("/companies/:id", s.requireUserSession, s.handleDeleteCompany)

	web.Get("/customers", s.requireUserSession, s.handleListCustomers)
	web.Get("/customers/:id", s.requireUserSession, s.handleGetCustomer)
	web.Post("/customers", s.requireUserSession, s.handleCreateCustomer)
	web.Put("/customers/:id", s.requireUserSession, s.handleUpdateCustomer)
	web.Delete("/customers/:id", s.requireUserSession, s.handleDeleteCustomer)

	web.Get("/frames", s.requireUserSession, s.handleListFrames)
	web.Get("/frames/:id", s.requireUserSession, s.handleGetFrame)
	web.Post("/frames", s.requireUserSession, s.handleCreateFrame)
	web.Put("/frames/:id", s.requireUserSession, s.handleUpdateFrame)
	web.Delete("/frames/:id", s.requireUserSession, s.handleDeleteFrame)

	web.Get("/templates", s.requireUserSession, s.handleListTemplate)
	web.Get("/templates/:id", s.requireUserSession, s.handleGetTemplate)
	web.Post("/templates", s.requireUserSession, s.requireAdmin, s.handleCreateTemplate)
	web.Put("/templates/:id", s.requireUserSession, s.requireAdmin, s.handleUpdateTemplate)
	web.Delete("/templates/:id", s.requireUserSession, s.requireAdmin, s.handleDeleteTemplate)

	web.Get("/render/:id", s.requireCompanySession, s.handleRenderDesign)

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

	web.Post("/subscriptions/products", s.requireUserSession, s.requireAdmin, s.handleCreateProduct)
	web.Get("/subscriptions/products", s.requireUserSession, s.requireAdmin, s.handleListProducts)
	web.Get("/subscriptions/plans", s.handleListPlan)
	web.Post("/subscriptions/plans", s.requireUserSession, s.requireAdmin, s.handleCreatePlan)
	web.Post("/subscriptions/plans/:id/subscribe", s.requireUserSession, s.handleSubscribeUser)
}
