# TYSMP Application Form

A full-stack web application for TYSMP (private vanilla Minecraft SMP) application submissions. This project includes:

- An ASP.NET Core backend with Entity Framework for data storage
- A Vue.js frontend with animations and a Minecraft-inspired design
- Form submission with validation
- Admin panel to view and export applications

## Features

- Interactive application form matching the required fields
- Beautiful animations using GSAP
- Minecraft-inspired design with custom styling
- Form validation and error handling
- Data storage in SQLite database
- Export applications to CSV
- Admin panel to view all submissions

## Setup and Run

### Prerequisites

- .NET 9.0 SDK
- Node.js and npm

### Backend Setup

1. Clone the repository
2. Navigate to the project root directory
3. Run the following commands:

```bash
dotnet restore
dotnet build
```

### Frontend Setup

1. Navigate to the ClientApp directory
2. Install dependencies:

```bash
npm install
```

### Run the Application

#### Development Mode

1. Run the backend:

```bash
dotnet run
```

2. In another terminal, run the frontend:

```bash
cd ClientApp
npm run dev
```

3. Access the application at:
   - Frontend: http://localhost:5173
   - Backend API: https://localhost:5001

#### Production Mode

To build the application for production:

1. Build the frontend:

```bash
cd ClientApp
npm run build
```

2. Run the .NET application:

```bash
dotnet run --environment Production
```

3. Access the application at https://localhost:5001

## Admin Access

The admin panel is available at `/admin` path. There is no authentication implemented in this demo version.

## License

This project is open-source and available under the MIT License. 