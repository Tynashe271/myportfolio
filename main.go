package main

import (
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type project struct {
	Title       string
	Description string
	Status      string
	Icon        string
	Tags        []string
	Highlights  []string
	Image       string
	URL         string
	AdminURL    string
	Repo        string
	Deployed    bool
}

type timelineEntry struct {
	Date, Title, Detail string
}

type testimonial struct {
	Quote, Author, Role string
}

type pageData struct {
	Name, Badge, Description                               string
	Phone, PhoneRaw                                        string
	EmailUniversity, EmailPersonal                         string
	GitHub, GitHubHandle                                   string
	Languages, Availability, Interests, Location, Learning string
	ResumeURL                                              string
	Projects                                               []project
	Timeline                                               []timelineEntry
	Testimonials                                           []testimonial
}

type contactRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
	SentAt  string `json:"sent_at"`
}

type application struct {
	template    *template.Template
	contactFile string
	contactMu   sync.Mutex
	deployments *deploymentLookup
}

func main() {
	app, err := newApplication()
	if err != nil {
		log.Fatal(err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	server := &http.Server{
		Addr:              "127.0.0.1:" + port,
		Handler:           app.routes(),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("Portfolio running at http://127.0.0.1:%s", port)
	log.Fatal(server.ListenAndServe())
}

func newApplication() (*application, error) {
	tmpl, err := template.New("index.html").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}).ParseFiles("web/index.html")
	if err != nil {
		return nil, err
	}
	return &application{template: tmpl, contactFile: filepath.Join("data", "contact-messages.jsonl"), deployments: newDeploymentLookup()}, nil
}

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", app.portfolio)
	mux.HandleFunc("POST /api/contact/", app.contact)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	return securityHeaders(mux)
}

func (app *application) portfolio(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := portfolioData()
	if _, err := os.Stat(filepath.Join("static", "resume.pdf")); err == nil {
		data.ResumeURL = "/static/resume.pdf"
	}
	app.deployments.updateProjects(r.Context(), data.Projects)
	if err := app.template.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Printf("render template: %v", err)
	}
}

func (app *application) contact(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()

	var message contactRequest
	if err := decoder.Decode(&message); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "Invalid request body."})
		return
	}
	message.Name = strings.TrimSpace(message.Name)
	message.Email = strings.TrimSpace(message.Email)
	message.Message = strings.TrimSpace(message.Message)
	if message.Name == "" || message.Email == "" || message.Message == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "Please fill all fields."})
		return
	}
	address, err := mail.ParseAddress(message.Email)
	if err != nil || address.Address != message.Email {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "Please enter a valid email address."})
		return
	}
	message.SentAt = time.Now().UTC().Format(time.RFC3339)
	if err := app.saveContact(message); err != nil {
		log.Printf("save contact: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": "Your message could not be saved. Please use the email link."})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "message": "Message saved successfully!"})
}

func (app *application) saveContact(message contactRequest) error {
	app.contactMu.Lock()
	defer app.contactMu.Unlock()
	if err := os.MkdirAll(filepath.Dir(app.contactFile), 0700); err != nil {
		return err
	}
	file, err := os.OpenFile(app.contactFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(message)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil && !errors.Is(err, http.ErrHandlerTimeout) {
		log.Printf("write response: %v", err)
	}
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

func portfolioData() pageData {
	return pageData{
		Name:        "Tinashe Nyenyesa",
		Badge:       "Industrial Attachment Candidate \u00b7 July 2026",
		Description: "Second-year Computer Science student at NUST (Part 2.2) with a strong passion for IoT, cyber security, and full-stack development. Certified in Cyber Security. Seeking an industrial attachment from July 2026 to contribute and grow with experienced teams.",
		Phone:       "+263 77 959 5732", PhoneRaw: "+263779595732",
		EmailUniversity: "tinashenyenyesa@students.nust.ac.zw", EmailPersonal: "tinashenyenyesa10@gmail.com",
		GitHub: "https://github.com/Tynashe271", GitHubHandle: "Tynashe271",
		Languages: "English (fluent) \u00b7 Shona (native)", Availability: "July 2026 \u2013 April 2027",
		Interests: "IoT, Software Engineering, Cyber Security, Embedded Systems",
		Location:  "Bulawayo, Zimbabwe \u00b7 Open to remote/hybrid opportunities", Learning: "Advanced JavaScript, Node.js, Cloud basics",
		Projects: []project{
			{Title: "NetworkGuard IDS", Description: "An authenticated AI-assisted intrusion detection system that classifies live traffic, CSV datasets, and packet captures, then surfaces alerts and model insights in a security dashboard.", Status: "Active", Tags: []string{"Python", "Flask", "Scikit-learn", "Cybersecurity"}, Highlights: []string{"Classifies live traffic, uploaded CSV datasets, and packet captures", "Security dashboard surfaces alerts and model insights in one place", "Authenticated access protects the detection tooling"}, URL: "https://www.tinashenyenyesatech.ac.zw", Deployed: true},
			{Title: "Bus Booking System", Description: "A multi-platform booking experience with a Laravel service layer, a polished Next.js web app, and a Flutter mobile client for passengers on the move.", Status: "In development", Tags: []string{"Laravel", "Next.js", "Flutter", "Mobile"}, URL: "http://localhost:3000", Repo: "C-Users-Vicario-Documents-Projects-bus-booking-system"},
			{Title: "HarvestHub Shop", Description: "A connected commerce ecosystem for agricultural products, pairing customer and administrator Flutter apps with a dedicated backend and operational workflows.", Status: "Live", Tags: []string{"Flutter", "Python", "E-commerce", "Mobile"}, URL: "https://harvesthub-grocery.web.app", Deployed: true},
			{Title: "Developer Portfolio", Description: "This project-first personal portfolio: a lightweight, responsive Go application designed to present experiments, products, and technical interests with personality.", Status: "Live", Tags: []string{"Go", "HTML", "CSS", "Responsive"}, URL: "/", Deployed: true},
			{Title: "ZB POS Pulse", Description: "A monitoring platform for POS availability, transaction performance, critical alerts, device heartbeats, and technician support-ticket workflows.", Status: "Built", Tags: []string{"Laravel", "Telemetry", "OpenAPI", "Operations"}, URL: "http://localhost:3000"},
			{Title: "Resolviax", Description: "A multilingual AI customer-support platform with evidence-grounded answers, human handoff, real-time conversations, ticket intelligence, automation, and privacy controls.", Status: "Active", Tags: []string{"Go", "Angular", "FastAPI", "RAG"}, URL: "http://localhost:4200"},
			{Title: "Restaurant Management System", Description: "An end-to-end restaurant platform covering menus, reservations, orders, payments, inventory, purchase orders, staff notifications, and reporting.", Status: "Built", Tags: []string{"Laravel", "REST API", "Payments", "Inventory"}, URL: "http://localhost:8000"},
			{Title: "Tinashe Messaging Platform", Description: "A privacy-first messaging platform with real-time delivery, Signal-Protocol identity, contacts, groups, calls, presence, media, moderation, and granular privacy tools.", Status: "Active", Tags: []string{"Laravel", "React", "WebSockets", "Encryption"}, URL: "http://localhost:5173"},
			{Title: "Tinashe V2", Description: "A polyglot rebuild of the messaging platform with a Spring Boot API, Go WebSocket relay, Next.js interface, PostgreSQL, and Redis-backed live events.", Status: "Phase 1 complete", Tags: []string{"Go", "Spring Boot", "Next.js", "PostgreSQL"}, URL: "http://localhost:3000"},
			{Title: "Anyschool University Portal", Description: "A full university platform for public, student, applicant, and administrator journeys, backed by a polyglot service architecture and private object storage.", Status: "Built", Tags: []string{"Nuxt", "Node.js", "Go", "PostgreSQL"}, URL: "http://localhost:3000", Repo: "university-portal"},
			{Title: "ZimMarket", Description: "A Zimbabwean online marketplace with product discovery, a live catalogue assistant, a Next.js storefront, and a scalable NestJS commerce backend.", Status: "Active", Tags: []string{"Next.js", "NestJS", "Prisma", "AI"}, URL: "http://localhost:3001", Repo: "ZimMarket"},
			{Title: "Maps Kayz Fashions", Description: "A deployed fashion-commerce platform with a Vue storefront, dedicated administration portal, native Expo shells, inventory and order workflows, and a NestJS API backed by PostgreSQL.", Status: "Live", Tags: []string{"Vue", "NestJS", "Expo", "PostgreSQL"}, URL: "https://shop.tinashenyenyesa.co.zw", Deployed: true},
			{Title: "GadgetHub", Description: "A production-focused electronics marketplace spanning customer and staff apps, catalogue search, inventory-aware checkout, payments, delivery, loyalty, trade-ins, repairs, support, and an AI shopping assistant.", Status: "Live", Tags: []string{"TypeScript", "React", "Prisma", "Flutter"}, Highlights: []string{"Separate customer and staff apps with catalogue search and inventory-aware checkout", "Payments, delivery, loyalty, trade-ins, and repair workflows in one system", "AI shopping assistant built into the customer experience"}, URL: "https://tynashe271.github.io/gadgethub/customer/", AdminURL: "https://tynashe271.github.io/gadgethub/admin/", Repo: "gadgethub", Deployed: true},
		},
		Timeline: []timelineEntry{
			{Date: "Right now", Title: "Building at NUST", Detail: "Part 2.2 of a Computer Science degree, learning by shipping the projects below."},
			{Date: "Jan–Mar 2026", Title: "Went deep on security", Detail: "Earned a Professional Certificate of Proficiency in Cyber Security from Lupane State University – Centre for Continuing Education (credential LSU-CCE02690)."},
			{Date: "Jul 2026 – Apr 2027", Title: "Open for an attachment", Detail: "Ready to join a team for the industrial attachment year and put all of this to work."},
		},
		Testimonials: []testimonial{},
	}
}
