'use client';

import { useState, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { MagnifyingGlassIcon, MapPinIcon, StarIcon, SparklesIcon, FireIcon, ChartBarIcon } from '@heroicons/react/24/outline';
import { venueService, geoService } from '@/lib/api';
import VenueCard from '@/components/VenueCard';
import type { VenueSearchParams } from '@/types';

export default function HomePage() {
  const [searchQuery, setSearchQuery] = useState('');
  const [searchParams, setSearchParams] = useState<Partial<VenueSearchParams>>({
    page: 1,
    limit: 12,
    sortBy: 'rating',
  });
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  // Search results query
  const { data: searchResults, isLoading: searchLoading } = useQuery({
    queryKey: ['venues', 'search', searchParams],
    queryFn: () => venueService.search(searchParams as VenueSearchParams),
    enabled: !!searchParams.query,
  });

  // Featured venues for homepage
  const { data: featuredVenues, isLoading: featuredLoading } = useQuery({
    queryKey: ['venues', 'featured'],
    queryFn: () => venueService.getFeatured(8),
    enabled: !searchParams.query,
  });

  const handleSearch = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!searchQuery.trim()) return;

    // Try to get user location
    let location;
    try {
      location = await geoService.getCurrentLocation();
    } catch {
      console.log('Location not available, searching without location');
    }

    const newParams: Partial<VenueSearchParams> = {
      query: searchQuery,
      latitude: location?.lat,
      longitude: location?.lng,
      radius: location ? 10 : undefined,
      page: 1,
      limit: 12,
      sortBy: 'rating',
    };
    
    setSearchParams(newParams);
  };

  const featuredCategories = [
    { name: '🍕 Restaurants', count: '2.4k+', trend: '+12%', emoji: '🍕' },
    { name: '🍸 Bars', count: '890+', trend: '+8%', emoji: '🍸' },
    { name: '☕ Cafes', count: '1.2k+', trend: '+15%', emoji: '☕' },
    { name: '🎵 Nightlife', count: '430+', trend: '+20%', emoji: '🎵' },
  ];

  const trendingSearches = ['Rooftop bars', 'Italian cuisine', 'Happy hour', 'Weekend brunch', 'Live music'];

  if (!mounted) return null;

  return (
    <div className="min-h-screen relative overflow-hidden">
      {/* Animated Background Elements */}
      <div className="absolute inset-0 overflow-hidden">
        <div className="absolute -top-40 -right-40 w-80 h-80 bg-purple-300 rounded-full mix-blend-multiply filter blur-xl opacity-70 animate-float"></div>
        <div className="absolute -bottom-40 -left-40 w-80 h-80 bg-yellow-300 rounded-full mix-blend-multiply filter blur-xl opacity-70 animate-float" style={{animationDelay: '2s'}}></div>
        <div className="absolute top-40 left-40 w-80 h-80 bg-pink-300 rounded-full mix-blend-multiply filter blur-xl opacity-70 animate-float" style={{animationDelay: '4s'}}></div>
      </div>

      {/* Modern Navigation */}
      <nav className="relative z-10 glass border-b border-white/20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center space-x-3">
              <SparklesIcon className="h-8 w-8 text-white animate-pulse-glow" />
              <span className="text-xl font-bold text-white gradient-text">VenueDiscovery</span>
            </div>
            <div className="hidden md:flex items-center space-x-8">
              <a href="#" className="text-white/80 hover:text-white transition-all duration-300 hover:scale-105">Explore</a>
              <a href="#" className="text-white/80 hover:text-white transition-all duration-300 hover:scale-105">Categories</a>
              <a href="#" className="text-white/80 hover:text-white transition-all duration-300 hover:scale-105">About</a>
              <button className="btn-modern">Get Started</button>
            </div>
          </div>
        </div>
      </nav>

      {/* Hero Section */}
      <main className="relative z-10">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-20 pb-16">
          <div className="text-center animate-slide-in-up">
            <h1 className="text-6xl md:text-8xl font-bold text-white mb-6 leading-tight">
              Discover
              <span className="block gradient-text animate-pulse-glow">Amazing Places</span>
            </h1>
            <p className="text-xl md:text-2xl text-white/80 mb-12 max-w-3xl mx-auto leading-relaxed">
              Find the perfect restaurants, bars, and cafes with AI-powered recommendations 
              and authentic community reviews ✨
            </p>

            {/* Ultra-Modern Search Bar */}
            <div className="max-w-2xl mx-auto mb-16">
              <form onSubmit={handleSearch} className="relative">
                <div className="glass-hover rounded-2xl p-2 animate-slide-in-up">
                  <div className="flex items-center">
                    <MagnifyingGlassIcon className="h-6 w-6 text-white/60 ml-4" />
                    <input
                      type="text"
                      value={searchQuery}
                      onChange={(e) => setSearchQuery(e.target.value)}
                      placeholder="Search for restaurants, bars, cafes..."
                      className="flex-1 bg-transparent border-none outline-none text-white placeholder-white/60 px-4 py-4 text-lg"
                    />
                    <button
                      type="submit"
                      className="btn-modern mr-2 animate-pulse-glow"
                      disabled={searchLoading}
                    >
                      {searchLoading ? (
                        <div className="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                      ) : (
                        'Search'
                      )}
                    </button>
                  </div>
                </div>
              </form>

              {/* Trending Searches */}
              <div className="mt-6 animate-fade-in">
                <p className="text-white/60 text-sm mb-3 flex items-center justify-center gap-2">
                  <FireIcon className="h-4 w-4" />
                  Trending searches:
                </p>
                <div className="flex flex-wrap justify-center gap-2">
                  {trendingSearches.map((search, index) => (
                    <button
                      key={index}
                      onClick={() => setSearchQuery(search)}
                      className="px-4 py-2 glass-hover rounded-full text-sm text-white/80 hover:text-white transition-all duration-300 hover:scale-105"
                    >
                      #{search}
                    </button>
                  ))}
                </div>
              </div>
            </div>

            {/* Animated Stats Cards */}
            <div className="grid grid-cols-2 md:grid-cols-4 gap-6 mb-20">
              {featuredCategories.map((category, index) => (
                <div
                  key={index}
                  className="card-modern text-center animate-slide-in-up hover:scale-105 transition-all duration-300"
                  style={{animationDelay: `${index * 0.1}s`}}
                >
                  <div className="text-4xl mb-3 animate-float" style={{animationDelay: `${index * 0.5}s`}}>
                    {category.emoji}
                  </div>
                  <div className="text-white text-2xl font-bold mb-1">{category.count}</div>
                  <div className="flex items-center justify-center text-green-400 text-sm">
                    <ChartBarIcon className="h-4 w-4 mr-1" />
                    {category.trend}
                  </div>
                  <div className="text-white/70 text-sm mt-1">{category.name.replace(/^\w+ /, '')}</div>
                </div>
              ))}
            </div>

            {/* Call to Action Buttons */}
            <div className="flex flex-col sm:flex-row gap-4 justify-center items-center animate-fade-in">
              <button className="btn-modern text-lg px-8 py-4 hover:scale-105 transition-all duration-300">
                <MapPinIcon className="h-5 w-5 mr-2 inline" />
                Find Near Me
              </button>
              <button className="glass-hover rounded-full px-8 py-4 text-white border border-white/20 hover:scale-105 transition-all duration-300">
                <StarIcon className="h-5 w-5 mr-2 inline" />
                Browse Categories
              </button>
            </div>
          </div>
        </div>

        {/* Search Results or Featured Venues */}
        <section className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pb-20">
          {searchParams.query ? (
            <div className="animate-slide-in-up">
              <div className="text-center mb-12">
                <h2 className="text-4xl font-bold text-white mb-4">
                  Results for &ldquo;{searchParams.query}&rdquo;
                </h2>
                {searchResults && (
                  <p className="text-white/70 text-lg">
                    {searchResults.pagination.total} amazing places found ✨
                  </p>
                )}
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
                {searchLoading
                  ? [...Array(6)].map((_, i) => (
                      <div key={i} className="card-modern animate-pulse">
                        <div className="h-48 bg-white/10 rounded-lg mb-4"></div>
                        <div className="h-4 bg-white/10 rounded mb-2"></div>
                        <div className="h-3 bg-white/10 rounded w-20"></div>
                      </div>
                    ))
                  : searchResults?.venues.map((venue, index) => (
                      <div
                        key={venue.id}
                        className="animate-slide-in-up"
                        style={{animationDelay: `${index * 0.1}s`}}
                      >
                        <VenueCard venue={venue} />
                      </div>
                    ))
                }
              </div>

              {searchResults && searchResults.venues.length === 0 && (
                <div className="text-center py-16">
                  <div className="text-6xl mb-4">🔍</div>
                  <p className="text-white/70 text-lg mb-4">No venues found for &ldquo;{searchParams.query}&rdquo;</p>
                  <button
                    onClick={() => {
                      setSearchQuery('');
                      setSearchParams({ page: 1, limit: 12, sortBy: 'rating' });
                    }}
                    className="btn-modern"
                  >
                    Clear search
                  </button>
                </div>
              )}
            </div>
          ) : (
            <div className="animate-slide-in-up">
              <h2 className="text-4xl font-bold text-white text-center mb-12">
                🌟 Featured Venues
              </h2>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
                {featuredLoading
                  ? [...Array(6)].map((_, i) => (
                      <div key={i} className="card-modern animate-pulse">
                        <div className="h-48 bg-white/10 rounded-lg mb-4"></div>
                        <div className="h-4 bg-white/10 rounded mb-2"></div>
                        <div className="h-3 bg-white/10 rounded w-20"></div>
                      </div>
                    ))
                  : featuredVenues?.map((venue, index) => (
                      <div
                        key={venue.id}
                        className="animate-slide-in-up"
                        style={{animationDelay: `${index * 0.1}s`}}
                      >
                        <VenueCard venue={venue} />
                      </div>
                    ))
                }
              </div>
            </div>
          )}
        </section>
      </main>

      {/* Modern Footer */}
      <footer className="relative z-10 glass border-t border-white/20 mt-20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
          <div className="text-center">
            <div className="flex items-center justify-center space-x-2 mb-4">
              <SparklesIcon className="h-6 w-6 text-white" />
              <span className="text-lg font-bold text-white">VenueDiscovery</span>
            </div>
            <p className="text-white/60 text-sm max-w-md mx-auto">
              Discover amazing places with AI-powered recommendations and authentic reviews from real people.
            </p>
            <div className="mt-6 text-white/40 text-xs">
              © 2024 VenueDiscovery. Made with ❤️ for food lovers.
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
}