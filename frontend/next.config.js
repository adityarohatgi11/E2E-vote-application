/** @type {import('next').NextConfig} */
const nextConfig = {
  images: {
    domains: ['localhost', 'example.com'], // Add your image domains
    remotePatterns: [
      {
        protocol: 'https',
        hostname: '**',
      },
    ],
  },
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: 'http://localhost:8080/v1/:path*', // Proxy to Go backend
      },
    ];
  },
  env: {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/v1',
    NEXT_PUBLIC_MAP_API_KEY: process.env.NEXT_PUBLIC_MAP_API_KEY || '',
  },
};

module.exports = nextConfig;
