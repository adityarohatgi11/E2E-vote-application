# VenueDiscovery Frontend

A clean, simple, and modern web application for discovering great venues. Built with Next.js, TypeScript, and Tailwind CSS.

## ✨ Simple & Modern Design

### Core Features
- **Clean Search** - Simple search with location detection
- **Venue Discovery** - Browse restaurants, bars, and cafes
- **Venue Details** - Essential information and reviews
- **Responsive Design** - Works perfectly on all devices
- **Fast Performance** - Optimized for speed and simplicity

### Design Philosophy
- **Minimal** - Clean, uncluttered interface
- **Intuitive** - Easy to use for everyone
- **Fast** - Quick loading and smooth interactions
- **Modern** - Contemporary design language
- **Accessible** - Works for all users

## Technology Stack

- **Framework**: Next.js 14 with App Router
- **Language**: TypeScript
- **Styling**: Tailwind CSS
- **Data**: TanStack Query (React Query)
- **Icons**: Heroicons
- **HTTP**: Axios

## Project Structure

```
frontend/
├── src/
│   ├── app/                 # Next.js App Router pages
│   │   ├── layout.tsx       # Root layout
│   │   ├── page.tsx         # Home page
│   │   ├── providers.tsx    # App providers
│   │   └── globals.css      # Global styles
│   ├── components/          # Reusable components
│   │   ├── VenueCard.tsx    # Venue display card
│   │   ├── SearchBar.tsx    # Search interface
│   │   └── ...              # Other components
│   ├── lib/                 # Utilities and configurations
│   │   ├── api.ts           # API client and services
│   │   └── utils.ts         # Helper functions
│   ├── types/               # TypeScript type definitions
│   │   └── index.ts         # API and UI types
│   └── hooks/               # Custom React hooks
├── public/                  # Static assets
├── package.json            # Dependencies and scripts
├── tailwind.config.js      # Tailwind configuration
├── next.config.js          # Next.js configuration
└── tsconfig.json           # TypeScript configuration
```

## 🚀 Quick Start

### Prerequisites
- Node.js 18+
- Go backend running on `http://localhost:8080`

### Setup (2 minutes)

```bash
# 1. Install dependencies
npm install

# 2. Set environment (optional)
cp env.example .env.local

# 3. Start development
npm run dev

# 4. Open http://localhost:3000
```

### Production Build
```bash
npm run build
npm start
```

## 🎯 Design Principles

### Simple & Clean
- **White backgrounds** - Clean, uncluttered look
- **Subtle borders** - Gentle gray-100 borders for definition
- **Minimal shadows** - Light shadow-sm for depth
- **Rounded corners** - Modern rounded-xl for cards

### Typography
- **Clear hierarchy** - Bold headings, medium body text
- **Good spacing** - Generous padding and margins
- **Readable sizes** - 16px+ for body text

### Colors
- **Gray scale** - Primary use of grays for simplicity
- **Blue accents** - Blue-600 for interactive elements
- **Yellow stars** - Classic yellow-400 for ratings
- **Subtle backgrounds** - Gray-50 for sections

## 🔧 Key Features

### Homepage
- **Hero search** - Large, centered search with rounded input
- **Category pills** - Simple category buttons with emojis
- **Venue grid** - Clean 4-column grid on desktop
- **Minimal header** - Just logo and navigation

### Venue Details
- **Hero image** - Full-width venue photo
- **Clean layout** - 2-column layout with sidebar
- **Contact info** - Easy-to-find contact details
- **Simple reviews** - Straightforward review display

### Search Results
- **Grid layout** - Consistent venue cards
- **Clear titles** - Easy-to-read venue names
- **Star ratings** - Visual rating display
- **Distance badges** - Location information

## 🚀 Performance

- **Minimal dependencies** - Only essential packages
- **Fast loading** - Optimized images and code
- **Smooth animations** - Subtle hover effects
- **Mobile first** - Responsive design

---

**Simple. Fast. Modern.** 🎯
