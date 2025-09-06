import './globals.css';
import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import { Providers } from './providers';

const inter = Inter({ subsets: ['latin'] });

export const metadata: Metadata = {
  title: 'VenueDiscovery - Find Amazing Places',
  description: 'Discover restaurants, bars, cafes and more with AI-powered recommendations and community reviews.',
  keywords: ['restaurants', 'venues', 'reviews', 'discovery', 'dining', 'bars', 'cafes'],
  authors: [{ name: 'VenueDiscovery Team' }],
  openGraph: {
    title: 'VenueDiscovery - Find Amazing Places',
    description: 'Discover restaurants, bars, cafes and more with AI-powered recommendations and community reviews.',
    type: 'website',
    locale: 'en_US',
    url: 'https://venuediscovery.com',
    siteName: 'VenueDiscovery',
  },
  twitter: {
    card: 'summary_large_image',
    title: 'VenueDiscovery - Find Amazing Places',
    description: 'Discover restaurants, bars, cafes and more with AI-powered recommendations and community reviews.',
  },
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      'max-video-preview': -1,
      'max-image-preview': 'large',
      'max-snippet': -1,
    },
  },
};

export const viewport = {
  width: 'device-width',
  initialScale: 1,
  themeColor: '#3b82f6',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <head>
        <link rel="icon" type="image/x-icon" href="/favicon.ico" />
        <link rel="apple-touch-icon" sizes="180x180" href="/apple-touch-icon.png" />
        <link rel="icon" type="image/png" sizes="32x32" href="/favicon-32x32.png" />
        <link rel="icon" type="image/png" sizes="16x16" href="/favicon-16x16.png" />
        <link rel="manifest" href="/site.webmanifest" />
      </head>
      <body className={inter.className}>
        <Providers>
          {children}
        </Providers>
      </body>
    </html>
  );
}
