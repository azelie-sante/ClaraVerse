import * as React from 'react';
import { useState, useEffect, useCallback } from 'react';
import './ImageCarousel.css';

interface ImageCarouselProps {
  images: string[];
  interval?: number;
  showDots?: boolean;
  rounded?: boolean;
  className?: string;
}

export const ImageCarousel: React.FC<ImageCarouselProps> = ({
  images,
  interval = 4000,
  showDots = true,
  rounded = true,
  className = '',
}) => {
  const [currentIndex, setCurrentIndex] = useState(0);
  const [isPaused, setIsPaused] = useState(false);

  const goToNext = useCallback(() => {
    setCurrentIndex(prevIndex => (prevIndex + 1) % images.length);
  }, [images.length]);

  const goToSlide = (index: number) => {
    setCurrentIndex(index);
  };

  useEffect(() => {
    if (isPaused || images.length <= 1) return;

    const timer = setInterval(goToNext, interval);
    return () => clearInterval(timer);
  }, [isPaused, interval, goToNext, images.length]);

  if (images.length === 0) return null;

  return (
    <div
      className={`image-carousel ${rounded ? 'image-carousel--rounded' : ''} ${className}`}
      onMouseEnter={() => setIsPaused(true)}
      onMouseLeave={() => setIsPaused(false)}
    >
      <div className="image-carousel__container">
        {images.map((image, index) => (
          <div
            key={index}
            className={`image-carousel__slide ${
              index === currentIndex ? 'image-carousel__slide--active' : ''
            }`}
          >
            <img
              src={image}
              alt={`Slide ${index + 1}`}
              className="image-carousel__image"
              loading={index === 0 ? 'eager' : 'lazy'}
            />
          </div>
        ))}
      </div>

      {showDots && images.length > 1 && (
        <div className="image-carousel__dots">
          {images.map((_, index) => (
            <button
              key={index}
              className={`image-carousel__dot ${
                index === currentIndex ? 'image-carousel__dot--active' : ''
              }`}
              onClick={() => goToSlide(index)}
              aria-label={`Go to slide ${index + 1}`}
            />
          ))}
        </div>
      )}
    </div>
  );
};
