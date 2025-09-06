import React from 'react';
import { MapPinIcon, HeartIcon, ShareIcon, EyeIcon } from '@heroicons/react/24/outline';
import { StarIcon as StarIconSolid, HeartIcon as HeartIconSolid } from '@heroicons/react/24/solid';
import type { Venue } from '@/types';

interface VenueCardProps {
  venue: Venue;
}

export default function VenueCard({ venue }: VenueCardProps) {
  const [isFavorited, setIsFavorited] = React.useState(false);
  const [isViewed, setIsViewed] = React.useState(false);

  const renderStars = (rating: number) => {
    return (
      <div className="flex items-center gap-1">
        {[...Array(5)].map((_, i) => (
          <StarIconSolid
            key={i}
            className={`h-4 w-4 ${i < Math.floor(rating) ? 'text-yellow-400' : 'text-white/30'}`}
          />
        ))}
        <span className="ml-2 text-sm text-white font-medium">{rating?.toFixed(1) || '0.0'}</span>
      </div>
    );
  };

  return (
    <div 
      className="group card-modern overflow-hidden relative cursor-pointer"
      onClick={() => setIsViewed(true)}
    >
      {/* Image Section */}
      <div className="relative h-52 overflow-hidden rounded-xl">
        {venue.coverImage ? (
          <img
            src={venue.coverImage}
            alt={venue.name}
            className="w-full h-full object-cover group-hover:scale-110 transition-transform duration-500"
          />
        ) : (
          <div className="w-full h-full bg-gradient-to-br from-purple-400 via-pink-500 to-red-500 flex items-center justify-center">
            <span className="text-white text-4xl opacity-80">
              {venue.category?.name?.charAt(0) || venue.name?.charAt(0) || '🏪'}
            </span>
          </div>
        )}
        
        {/* Gradient Overlay */}
        <div className="absolute inset-0 bg-gradient-to-t from-black/60 via-transparent to-transparent"></div>
        
        {/* Distance Badge */}
        {venue.distance && (
          <div className="absolute top-3 left-3 glass rounded-full px-3 py-1 text-white text-xs font-medium">
            <MapPinIcon className="h-3 w-3 inline mr-1" />
            {venue.distance.toFixed(1)}km
          </div>
        )}
        
        {/* Featured Badge */}
        {venue.isFeatured && (
          <div className="absolute top-3 left-3 bg-gradient-to-r from-yellow-400 to-orange-500 rounded-full px-3 py-1 text-white text-xs font-bold">
            ⭐ Featured
          </div>
        )}
        
        {/* Action Buttons */}
        <div className="absolute top-3 right-3 flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity duration-300">
          <button
            onClick={(e) => {
              e.stopPropagation();
              setIsFavorited(!isFavorited);
            }}
            className="glass rounded-full p-2 hover:scale-110 transition-transform duration-200"
          >
            {isFavorited ? (
              <HeartIconSolid className="h-4 w-4 text-red-400" />
            ) : (
              <HeartIcon className="h-4 w-4 text-white" />
            )}
          </button>
          <button 
            onClick={(e) => e.stopPropagation()}
            className="glass rounded-full p-2 hover:scale-110 transition-transform duration-200"
          >
            <ShareIcon className="h-4 w-4 text-white" />
          </button>
          <button 
            onClick={(e) => e.stopPropagation()}
            className="glass rounded-full p-2 hover:scale-110 transition-transform duration-200"
          >
            <EyeIcon className="h-4 w-4 text-white" />
          </button>
        </div>
        
        {/* Rating Badge */}
        <div className="absolute bottom-3 left-3">
          {renderStars(venue.averageRating || 0)}
        </div>
        
        {/* Price Range */}
        {venue.priceRange && (
          <div className="absolute bottom-3 right-3 glass rounded-full px-3 py-1">
            <span className="text-white text-sm font-medium">{venue.priceRange}</span>
          </div>
        )}
      </div>
      
      {/* Content Section */}
      <div className="p-6">
        <div className="flex items-start justify-between mb-3">
          <h3 className="font-bold text-white text-lg leading-tight group-hover:text-purple-300 transition-colors duration-300">
            {venue.name}
          </h3>
        </div>
        
        {/* Category */}
        {venue.category && (
          <div className="inline-flex items-center gap-2 mb-3">
            <span className="text-purple-300 text-sm font-medium">
              {venue.category.name}
            </span>
          </div>
        )}
        
        {/* Description */}
        {(venue.shortDescription || venue.description) && (
          <p className="text-white/70 text-sm mb-4 line-clamp-2 leading-relaxed">
            {venue.shortDescription || venue.description}
          </p>
        )}
        
        {/* Address */}
        <div className="flex items-center text-white/60 text-sm mb-4">
          <MapPinIcon className="h-4 w-4 mr-2 flex-shrink-0" />
          <span className="truncate">{venue.address}</span>
        </div>
        
        {/* Amenities Tags */}
        {venue.amenities && Array.isArray(venue.amenities) && venue.amenities.length > 0 && (
          <div className="flex flex-wrap gap-1 mb-4">
            {venue.amenities.slice(0, 3).map((amenity, index) => (
              <span
                key={index}
                className="text-xs px-2 py-1 glass rounded-full text-white/80"
              >
                {amenity}
              </span>
            ))}
            {venue.amenities.length > 3 && (
              <span className="text-xs px-2 py-1 glass rounded-full text-white/80">
                +{venue.amenities.length - 3} more
              </span>
            )}
          </div>
        )}
        
        {/* Stats Row */}
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-4 text-xs text-white/60">
            {venue.totalRatings && (
              <span>{venue.totalRatings} reviews</span>
            )}
            {venue.averageCostPerPerson && (
              <span>${venue.averageCostPerPerson}/person</span>
            )}
          </div>
          {venue.isOpen !== undefined && (
            <span className={`text-xs font-medium px-2 py-1 rounded-full ${
              venue.isOpen 
                ? 'bg-green-500/20 text-green-400' 
                : 'bg-red-500/20 text-red-400'
            }`}>
              {venue.isOpen ? '🟢 Open' : '🔴 Closed'}
            </span>
          )}
        </div>
        
        {/* Action Footer */}
        <div className="flex items-center justify-between pt-4 border-t border-white/10">
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 bg-green-400 rounded-full animate-pulse"></div>
            <span className="text-white/60 text-xs">Available</span>
          </div>
          <button className="btn-modern text-sm px-4 py-2 opacity-0 group-hover:opacity-100 transition-all duration-300 hover:scale-105">
            View Details →
          </button>
        </div>
      </div>
      
      {/* Hover Glow Effect */}
      <div className="absolute inset-0 rounded-2xl opacity-0 group-hover:opacity-100 transition-opacity duration-500 pointer-events-none">
        <div className="absolute inset-0 rounded-2xl bg-gradient-to-r from-purple-400/20 via-pink-400/20 to-red-400/20 blur-xl"></div>
      </div>
      
      {/* Click Ripple Effect */}
      {isViewed && (
        <div className="absolute inset-0 rounded-2xl bg-white/10 animate-ping pointer-events-none"></div>
      )}
    </div>
  );
}