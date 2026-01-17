# Instagram Lite Frontend

A simple React application built with Vite and Tailwind CSS.

## Technologies Used

- **React** - UI library
- **Vite** - Build tool and dev server
- **Tailwind CSS** - Utility-first CSS framework

## Getting Started

### Prerequisites

- Node.js (v18 or higher recommended)
- npm

### Installation

1. Navigate to the frontend directory:
```bash
cd frontend
```

2. Install dependencies:
```bash
npm install
```

### Running the Development Server

```bash
npm run dev
```

The app will be available at `http://localhost:5173`

### Building for Production

```bash
npm run build
```

The production-ready files will be in the `dist` directory.

### Preview Production Build

```bash
npm run preview
```

## Project Structure

```
frontend/
├── public/          # Static assets
├── src/
│   ├── App.jsx      # Main app component
│   ├── App.css      # App styles
│   ├── index.css    # Global styles with Tailwind directives
│   └── main.jsx     # App entry point
├── index.html       # HTML template
├── tailwind.config.js
├── postcss.config.js
└── package.json
```
