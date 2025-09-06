'use client';

import { useQuery } from '@tanstack/react-query';
import { useParams, useRouter } from 'next/navigation';
import { ArrowLeftIcon, MapPinIcon, PhoneIcon, GlobeAltIcon } from '@heroicons/react/24/outline';
import { StarIcon as StarSolid } from '@heroicons/react/24/solid';
import { venueService, reviewService } from '@/lib/api';

export default function VenueDetailPage() {
  const params = useParams();
  const router = useRouter();
  const venueId = parseInt(params.id as string);

  // Get venue details
  const { data: venue, isLoading: venueLoading } = useQuery({
    queryKey: ['venue', venueId],
    queryFn: () => venueService.getById(venueId),
    enabled: !!venueId,
  });

  // Get venue reviews
  const { data: reviewsData, isLoading: reviewsLoading } = useQuery({
    queryKey: ['reviews', venueId],
    queryFn: () => reviewService.getVenueReviews(venueId, { limit: 6 }),
    enabled: !!venueId,
  });

  const renderStars = (rating: number, size = 'h-4 w-4') => {
    return (
      <div className="flex items-center gap-1">
        {[...Array(5)].map((_, i) => (
          <StarSolid
            key={i}
            className={`${size} ${i < Math.floor(rating) ? 'text-yellow-400' : 'text-gray-300'}`}
          />
        ))}
        <span className="text-sm text-gray-600 ml-1">{rating.toFixed(1)}</span>
      </div>
    );
  };

  if (venueLoading) {
    return (
      <div className="min-h-screen bg-white">
        <div className="max-w-4xl mx-auto px-4 py-8">
          <div className="animate-pulse">
            <div className="h-8 bg-gray-200 rounded w-32 mb-6"></div>
            <div className="aspect-[16/9] bg-gray-200 rounded-xl mb-6"></div>
            <div className="h-8 bg-gray-200 rounded w-64 mb-4"></div>
            <div className="h-4 bg-gray-200 rounded w-48 mb-6"></div>
            <div className="grid md:grid-cols-3 gap-8">
              <div className="md:col-span-2">
                <div className="h-6 bg-gray-200 rounded w-32 mb-4"></div>
                <div className="space-y-3">
                  <div className="h-4 bg-gray-200 rounded"></div>
                  <div className="h-4 bg-gray-200 rounded w-3/4"></div>
                </div>
              </div>
              <div>
                <div className="h-6 bg-gray-200 rounded w-24 mb-4"></div>
                <div className="space-y-2">
                  <div className="h-4 bg-gray-200 rounded"></div>
                  <div className="h-4 bg-gray-200 rounded w-2/3"></div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (!venue) {
    return (
      <div className="min-h-screen bg-white flex items-center justify-center">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-gray-900 mb-2">Venue not found</h1>
          <p className="text-gray-600 mb-4">The venue you&rsquo;re looking for doesn&rsquo;t exist.</p>
          <button
            onClick={() => router.push('/')}
            className="text-blue-600 hover:text-blue-700 font-medium"
          >
            Go back home
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-white">
      {/* Header */}
      <header className="border-b border-gray-100">
        <div className="max-w-6xl mx-auto px-4 py-4">
          <button
            onClick={() => router.back()}
            className="flex items-center gap-2 text-gray-600 hover:text-gray-900 transition-colors"
          >
            <ArrowLeftIcon className="h-5 w-5" />
            Back
          </button>
        </div>
      </header>

      <main className="max-w-4xl mx-auto px-4 py-8">
        {/* Hero Image */}
        {venue.venue.coverImage && (
          <div className="aspect-[16/9] rounded-xl overflow-hidden mb-8">
            <img
              src={venue.venue.coverImage}
              alt={venue.venue.name}
              className="w-full h-full object-cover"
            />
          </div>
        )}

        {/* Main Content */}
        <div className="grid md:grid-cols-3 gap-8">
          {/* Left Column - Main Info */}
          <div className="md:col-span-2">
            {/* Title & Rating */}
            <div className="mb-6">
              <div className="flex items-start justify-between mb-2">
                <h1 className="text-3xl font-bold text-gray-900">{venue.venue.name}</h1>
                {venue.venue.priceRange && (
                  <span className="text-lg text-gray-500">{venue.venue.priceRange}</span>
                )}
              </div>
              
              <p className="text-gray-600 mb-3">{venue.venue.category?.name}</p>
              
              <div className="flex items-center gap-4">
                {renderStars(venue.venue.averageRating, 'h-5 w-5')}
                <span className="text-gray-600">
                  {venue.venue.totalReviews} review{venue.venue.totalReviews !== 1 ? 's' : ''}
                </span>
              </div>
            </div>

            {/* Description */}
            {venue.venue.description && (
              <div className="mb-8">
                <h2 className="text-xl font-semibold text-gray-900 mb-3">About</h2>
                <p className="text-gray-700 leading-relaxed">{venue.venue.description}</p>
              </div>
            )}

            {/* Amenities */}
            {venue.venue.amenities && venue.venue.amenities.length > 0 && (
              <div className="mb-8">
                <h2 className="text-xl font-semibold text-gray-900 mb-3">Amenities</h2>
                <div className="flex flex-wrap gap-2">
                  {venue.venue.amenities.map((amenity, index) => (
                    <span
                      key={index}
                      className="bg-gray-100 text-gray-700 px-3 py-1 rounded-full text-sm"
                    >
                      {amenity}
                    </span>
                  ))}
                </div>
              </div>
            )}

            {/* Reviews */}
            <div>
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-xl font-semibold text-gray-900">Reviews</h2>
                <button className="text-blue-600 hover:text-blue-700 font-medium text-sm">
                  Write a review
                </button>
              </div>

              {reviewsLoading ? (
                <div className="space-y-4">
                  {[...Array(3)].map((_, i) => (
                    <div key={i} className="animate-pulse p-4 border border-gray-100 rounded-lg">
                      <div className="flex items-center gap-3 mb-3">
                        <div className="h-8 w-8 bg-gray-200 rounded-full"></div>
                        <div>
                          <div className="h-4 bg-gray-200 rounded w-24 mb-1"></div>
                          <div className="h-3 bg-gray-200 rounded w-16"></div>
                        </div>
                      </div>
                      <div className="h-4 bg-gray-200 rounded mb-2"></div>
                      <div className="h-4 bg-gray-200 rounded w-3/4"></div>
                    </div>
                  ))}
                </div>
              ) : reviewsData && reviewsData.reviews.length > 0 ? (
                <div className="space-y-4">
                  {reviewsData.reviews.map((review) => (
                    <div key={review.id} className="p-4 border border-gray-100 rounded-lg">
                      <div className="flex items-center gap-3 mb-3">
                        <div className="h-8 w-8 bg-gray-200 rounded-full flex items-center justify-center">
                          <span className="text-sm font-medium text-gray-600">
                            {review.userName?.charAt(0) || 'U'}
                          </span>
                        </div>
                        <div>
                          <p className="font-medium text-gray-900">{review.userName || 'Anonymous'}</p>
                          <div className="flex items-center gap-2">
                            {renderStars(review.overallRating)}
                            <span className="text-xs text-gray-500">
                              {new Date(review.createdAt).toLocaleDateString()}
                            </span>
                          </div>
                        </div>
                      </div>
                      
                      {review.title && (
                        <h4 className="font-medium text-gray-900 mb-2">{review.title}</h4>
                      )}
                      
                      {review.reviewText && (
                        <p className="text-gray-700 text-sm leading-relaxed">{review.reviewText}</p>
                      )}
                    </div>
                  ))}
                  
                  {venue.venue.totalReviews > 6 && (
                    <button className="w-full py-3 text-blue-600 hover:text-blue-700 font-medium border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors">
                      View all {venue.venue.totalReviews} reviews
                    </button>
                  )}
                </div>
              ) : (
                <div className="text-center py-8 border border-gray-100 rounded-lg">
                  <p className="text-gray-500">No reviews yet</p>
                  <button className="mt-2 text-blue-600 hover:text-blue-700 font-medium">
                    Be the first to review
                  </button>
                </div>
              )}
            </div>
          </div>

          {/* Right Column - Details */}
          <div>
            {/* Contact Info */}
            <div className="bg-gray-50 rounded-lg p-6 mb-6">
              <h3 className="font-semibold text-gray-900 mb-4">Contact</h3>
              
              <div className="space-y-3">
                <div className="flex items-start gap-3">
                  <MapPinIcon className="h-5 w-5 text-gray-400 mt-0.5" />
                  <div>
                    <p className="text-sm text-gray-900">{venue.venue.address}</p>
                    {venue.venue.city && (
                      <p className="text-xs text-gray-500">
                        {venue.venue.city.name}, {venue.venue.city.state}
                      </p>
                    )}
                  </div>
                </div>

                {venue.venue.phone && (
                  <div className="flex items-center gap-3">
                    <PhoneIcon className="h-5 w-5 text-gray-400" />
                    <a
                      href={`tel:${venue.venue.phone}`}
                      className="text-sm text-blue-600 hover:text-blue-700"
                    >
                      {venue.venue.phone}
                    </a>
                  </div>
                )}

                {venue.venue.website && (
                  <div className="flex items-center gap-3">
                    <GlobeAltIcon className="h-5 w-5 text-gray-400" />
                    <a
                      href={venue.venue.website}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-sm text-blue-600 hover:text-blue-700"
                    >
                      Visit website
                    </a>
                  </div>
                )}
              </div>
            </div>

            {/* Quick Stats */}
            <div className="bg-gray-50 rounded-lg p-6">
              <h3 className="font-semibold text-gray-900 mb-4">Quick Info</h3>
              
              <div className="space-y-3">
                <div className="flex justify-between">
                  <span className="text-sm text-gray-600">Rating</span>
                  <span className="text-sm font-medium text-gray-900">
                    {venue.venue.averageRating.toFixed(1)}/5.0
                  </span>
                </div>
                
                <div className="flex justify-between">
                  <span className="text-sm text-gray-600">Reviews</span>
                  <span className="text-sm font-medium text-gray-900">
                    {venue.venue.totalReviews}
                  </span>
                </div>

                {venue.venue.averageCostPerPerson && (
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-600">Avg. cost</span>
                    <span className="text-sm font-medium text-gray-900">
                      ${venue.venue.averageCostPerPerson}
                    </span>
                  </div>
                )}

                {venue.venue.priceRange && (
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-600">Price range</span>
                    <span className="text-sm font-medium text-gray-900">
                      {venue.venue.priceRange}
                    </span>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}
