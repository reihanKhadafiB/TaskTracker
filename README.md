# Task Tracker

A full-stack, premium Task Management application built with **Golang** (Backend) and **Astro + React** (Frontend). Designed with a clean, modern, and responsive user interface, featuring contextual task grouping, subtasks, and PDF report generation.

## 🌟 Features

- **Authentication**: Secure JWT-based Login and Registration.
- **Context Management**: Group your tasks into customizable Contexts (e.g., Work, Personal) with color coding.
- **Task & Subtask System**: 
  - Create, read, update, and delete tasks.
  - Break down tasks into smaller, manageable subtasks.
  - Mark tasks and subtasks as completed.
- **PDF Export**: Generate and download a comprehensive PDF report of all tasks within a specific Context.
- **Premium UI/UX**: Built with Vanilla CSS utilizing modern web design principles like Glassmorphism, CSS Variables, and smooth micro-animations.

## 🛠 Tech Stack

**Backend:**
- [Go (Golang)](https://golang.org/) - API Server
- [PostgreSQL](https://www.postgresql.org/) - Database
- [pgx](https://github.com/jackc/pgx) - PostgreSQL driver and toolkit
- [golang-jwt](https://github.com/golang-jwt/jwt) - JWT Authentication
- [go-pdf/fpdf](https://github.com/go-pdf/fpdf) - PDF Generation

**Frontend:**
- [Astro](https://astro.build/) - Web Framework
- [React](https://reactjs.org/) - UI Components
- Vanilla CSS - Styling and Animations

## 📋 Prerequisites

Make sure you have the following installed on your machine:
- **Go** (v1.22 or newer)
- **Node.js** (v22.12.0 or newer)
- **PostgreSQL** (v15 or newer)
- [golang-migrate](https://github.com/golang-migrate/migrate) (for database migrations)

## 🚀 Getting Started

### 1. Clone the Repository
```bash
git clone https://github.com/reihanKhadafiB/TaskTracker.git
cd TaskTracker
```

### 2. Database Setup
Create a new PostgreSQL database for the project:
```sql
CREATE DATABASE tasktracker;
```

Run the database migrations to create the required tables:
```bash
migrate -path migrations -database "postgres://postgres:password@localhost:5432/tasktracker?sslmode=disable" up
```
*(Adjust the database URL according to your PostgreSQL credentials)*

### 3. Environment Variables
Create a `.env` file in the root directory and configure the following variables:

```ini
PORT=8080
DATABASE_URL=postgres://postgres:password@localhost:5432/tasktracker?sslmode=disable
JWT_SECRET=your_super_secret_jwt_key
PUBLIC_API_BASE_URL=http://localhost:8080
```

### 4. Create an Initial User (Seeder)
To log into the application, you can generate an initial user account by running the seeder script:
```bash
go run cmd/seed/main.go
```
By default, this creates a user based on the credentials defined in `cmd/seed/main.go`. You can edit that file before running the script if you wish to change the default email and password.

---

## 💻 Running the Application

### Start the Backend Server
Open a terminal in the root directory and run:
```bash
go run cmd/api/main.go
```
The backend API will start on `http://localhost:8080`.

### Start the Frontend Web App
Open a **new** terminal, navigate to the frontend directory, and start the development server:
```bash
cd frontend
npm install
npm run dev
```
The frontend application will be available at `http://localhost:4321`.

## 📂 Project Structure

```
.
├── cmd/
│   ├── api/          # Main backend application entrypoint
│   └── seed/         # Database seeder script
├── frontend/         # Astro & React frontend application
│   ├── src/          # Frontend source code (components, pages, lib, styles)
│   └── public/       # Static frontend assets
├── internal/         # Backend internal packages
│   ├── handler/      # HTTP handlers and controllers
│   ├── middleware/   # HTTP middlewares (CORS, Auth)
│   ├── model/        # Database models and structs
│   ├── repository/   # Database access layer
│   ├── router/       # HTTP routing definitions
│   └── service/      # Business logic layer
└── migrations/       # SQL migration files
```

## 📝 License

This project is open-source and available under the MIT License.
