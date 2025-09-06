'use client';

import { useState, useRef, useEffect } from 'react';
import { MagnifyingGlassIcon, MapPinIcon, AdjustmentsHorizontalIcon } from '@heroicons/react/24/outline';
import { geoService } from '@/lib/api';

interface SearchBarProps {
  onSearch: (query: string, location?: { lat: number; lng: number }) => void;
  onFiltersToggle: () => void;
  placeholder?: string;
  showLocationButton?: boolean;
  showFiltersButton?: boolean;
  className?: string;
}

export default function SearchBar({
  onSearch,
  onFiltersToggle,
  placeholder = "Search restaurants, bars, cafes...",
  showLocationButton = true,
  showFiltersButton = true,
  className = "",
}: SearchBarProps) {
  const [query, setQuery] = useState('');
  const [isLocationLoading, setIsLocationLoading] = useState(false);
  const [locationError, setLocationError] = useState<string | null>(null);
  const [currentLocation, setCurrentLocation] = useState<{ lat: number; lng: number } | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSearch(query, currentLocation || undefined);
  };

  const handleLocationClick = async () => {
    if (currentLocation) {
      // If we already have location, use it
      onSearch(query, currentLocation);
      return;
    }

    setIsLocationLoading(true);
    setLocationError(null);

    try {
      const location = await geoService.getCurrentLocation();
      setCurrentLocation(location);
      onSearch(query, location);
    } catch (error) {
      console.error('Failed to get location:', error);
      setLocationError('Unable to get your location');
    } finally {
      setIsLocationLoading(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSubmit(e);
    }
  };

  useEffect(() => {
    // Auto-focus on mount
    if (inputRef.current) {
      inputRef.current.focus();
    }
  }, []);

  return (
    <div className={`w-full max-w-4xl mx-auto ${className}`}>
      <form onSubmit={handleSubmit} className="relative">
        <div className="relative flex items-center">
          {/* Search Input */}
          <div className="relative flex-1">
            <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
              <MagnifyingGlassIcon className="h-5 w-5 text-gray-400" />
            </div>
            <input
              ref={inputRef}
              type="text"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              onKeyDown={handleKeyDown}
              className="block w-full pl-10 pr-3 py-3 border border-gray-300 rounded-l-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent bg-white text-gray-900 placeholder-gray-500 text-base"
              placeholder={placeholder}
            />
          </div>

          {/* Location Button */}
          {showLocationButton && (
            <button
              type="button"
              onClick={handleLocationClick}
              disabled={isLocationLoading}
              className={`px-4 py-3 border-t border-b border-gray-300 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:z-10 transition-colors ${
                isLocationLoading ? 'cursor-not-allowed opacity-50' : ''
              } ${currentLocation ? 'text-primary-600' : 'text-gray-500'}`}
              title={currentLocation ? 'Using your location' : 'Use my location'}
            >
              {isLocationLoading ? (
                <div className="animate-spin h-5 w-5 border-2 border-gray-300 border-t-primary-600 rounded-full"></div>
              ) : (
                <MapPinIcon className={`h-5 w-5 ${currentLocation ? 'text-primary-600' : 'text-gray-400'}`} />
              )}
            </button>
          )}

          {/* Filters Button */}
          {showFiltersButton && (
            <button
              type="button"
              onClick={onFiltersToggle}
              className="px-4 py-3 border border-gray-300 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:z-10 transition-colors"
              title="Filters"
            >
              <AdjustmentsHorizontalIcon className="h-5 w-5 text-gray-400" />
            </button>
          )}

          {/* Search Button */}
          <button
            type="submit"
            className="px-6 py-3 bg-primary-600 text-white font-medium rounded-r-lg hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 transition-colors"
          >
            Search
          </button>
        </div>
      </form>

      {/* Location Error */}
      {locationError && (
        <div className="mt-2 text-sm text-red-600 bg-red-50 border border-red-200 rounded-md px-3 py-2">
          {locationError}
          <button
            onClick={() => setLocationError(null)}
            className="ml-2 text-red-800 hover:text-red-900 font-medium"
          >
            Dismiss
          </button>
        </div>
      )}

      {/* Location Success */}
      {currentLocation && !locationError && (
        <div className="mt-2 text-sm text-green-600 bg-green-50 border border-green-200 rounded-md px-3 py-2 flex items-center justify-between">
          <span className="flex items-center gap-1">
            <MapPinIcon className="h-4 w-4" />
            Using your current location
          </span>
          <button
            onClick={() => setCurrentLocation(null)}
            className="text-green-800 hover:text-green-900 font-medium"
          >
            Clear
          </button>
        </div>
      )}

      {/* Search Suggestions (placeholder for future implementation) */}
      {query && query.length > 2 && (
        <div className="absolute top-full left-0 right-0 mt-1 bg-white border border-gray-200 rounded-lg shadow-lg z-50">
          {/* This would be populated with search suggestions from the API */}
          <div className="p-2 text-sm text-gray-500">
            Press Enter to search for &ldquo;{query}&rdquo;
          </div>
        </div>
      )}
    </div>
  );
}
