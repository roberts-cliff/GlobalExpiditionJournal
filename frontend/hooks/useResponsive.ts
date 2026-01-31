import { useState, useEffect } from 'react';
import { Dimensions, ScaledSize } from 'react-native';

export type DeviceType = 'phone' | 'tablet' | 'desktop';
export type Orientation = 'portrait' | 'landscape';

interface ResponsiveInfo {
  width: number;
  height: number;
  deviceType: DeviceType;
  orientation: Orientation;
  isPhone: boolean;
  isTablet: boolean;
  isDesktop: boolean;
  isPortrait: boolean;
  isLandscape: boolean;
}

// Breakpoints based on common device sizes
const PHONE_MAX_WIDTH = 480;
const TABLET_MAX_WIDTH = 1024;

function getDeviceType(width: number): DeviceType {
  if (width <= PHONE_MAX_WIDTH) return 'phone';
  if (width <= TABLET_MAX_WIDTH) return 'tablet';
  return 'desktop';
}

function getOrientation(width: number, height: number): Orientation {
  return width >= height ? 'landscape' : 'portrait';
}

function calculateResponsiveInfo(dimensions: ScaledSize): ResponsiveInfo {
  const { width, height } = dimensions;
  const deviceType = getDeviceType(width);
  const orientation = getOrientation(width, height);

  return {
    width,
    height,
    deviceType,
    orientation,
    isPhone: deviceType === 'phone',
    isTablet: deviceType === 'tablet',
    isDesktop: deviceType === 'desktop',
    isPortrait: orientation === 'portrait',
    isLandscape: orientation === 'landscape',
  };
}

export function useResponsive(): ResponsiveInfo {
  const [dimensions, setDimensions] = useState(() => Dimensions.get('window'));

  useEffect(() => {
    const subscription = Dimensions.addEventListener('change', ({ window }) => {
      setDimensions(window);
    });

    return () => subscription.remove();
  }, []);

  return calculateResponsiveInfo(dimensions);
}

// Utility hook for responsive values
export function useResponsiveValue<T>(values: {
  phone?: T;
  tablet?: T;
  desktop?: T;
  default: T;
}): T {
  const { deviceType } = useResponsive();

  if (deviceType === 'phone' && values.phone !== undefined) {
    return values.phone;
  }
  if (deviceType === 'tablet' && values.tablet !== undefined) {
    return values.tablet;
  }
  if (deviceType === 'desktop' && values.desktop !== undefined) {
    return values.desktop;
  }

  return values.default;
}

// Utility hook for responsive spacing/sizing
export function useResponsiveSpacing(): {
  small: number;
  medium: number;
  large: number;
  padding: number;
  margin: number;
} {
  const { isPhone, isTablet } = useResponsive();

  if (isPhone) {
    return {
      small: 4,
      medium: 8,
      large: 16,
      padding: 12,
      margin: 8,
    };
  }

  if (isTablet) {
    return {
      small: 6,
      medium: 12,
      large: 24,
      padding: 16,
      margin: 12,
    };
  }

  // Desktop
  return {
    small: 8,
    medium: 16,
    large: 32,
    padding: 24,
    margin: 16,
  };
}

// Utility hook for responsive font sizes
export function useResponsiveFontSize(): {
  small: number;
  body: number;
  title: number;
  header: number;
} {
  const { isPhone, isTablet } = useResponsive();

  if (isPhone) {
    return {
      small: 12,
      body: 14,
      title: 18,
      header: 24,
    };
  }

  if (isTablet) {
    return {
      small: 13,
      body: 16,
      title: 22,
      header: 28,
    };
  }

  // Desktop
  return {
    small: 14,
    body: 16,
    title: 24,
    header: 32,
  };
}

// Utility hook for number of columns in grid layouts
export function useResponsiveColumns(baseColumns: number = 2): number {
  const { isPhone, isTablet } = useResponsive();

  if (isPhone) return Math.max(1, baseColumns - 1);
  if (isTablet) return baseColumns;
  return baseColumns + 1;
}
