import React, { useMemo } from 'react';
import { View, Text, StyleSheet, TouchableOpacity, ScrollView } from 'react-native';
import { Country, Visit } from '../types/api';

interface RegionData {
  name: string;
  emoji: string;
  color: string;
  visitedColor: string;
}

const REGION_CONFIG: Record<string, RegionData> = {
  'Europe': {
    name: 'Europe',
    emoji: '🏰',
    color: '#e0e7ff',
    visitedColor: '#6366f1',
  },
  'Asia': {
    name: 'Asia',
    emoji: '🏯',
    color: '#fce7f3',
    visitedColor: '#ec4899',
  },
  'Africa': {
    name: 'Africa',
    emoji: '🦁',
    color: '#fef3c7',
    visitedColor: '#f59e0b',
  },
  'Americas': {
    name: 'Americas',
    emoji: '🗽',
    color: '#d1fae5',
    visitedColor: '#10b981',
  },
  'Oceania': {
    name: 'Oceania',
    emoji: '🏝️',
    color: '#e0f2fe',
    visitedColor: '#0ea5e9',
  },
};

interface WorldMapProps {
  countries: Country[];
  visitedCountryIds: Set<number>;
  onRegionPress: (region: string) => void;
  onCountryPress: (country: Country) => void;
}

export function WorldMap({
  countries,
  visitedCountryIds,
  onRegionPress,
  onCountryPress,
}: WorldMapProps) {
  // Calculate stats per region
  const regionStats = useMemo(() => {
    const stats: Record<string, { total: number; visited: number; countries: Country[] }> = {};

    Object.keys(REGION_CONFIG).forEach((region) => {
      stats[region] = { total: 0, visited: 0, countries: [] };
    });

    countries.forEach((country) => {
      if (stats[country.region]) {
        stats[country.region].total++;
        stats[country.region].countries.push(country);
        if (visitedCountryIds.has(country.id)) {
          stats[country.region].visited++;
        }
      }
    });

    return stats;
  }, [countries, visitedCountryIds]);

  // Get recently visited countries (up to 5)
  const recentlyVisited = useMemo(() => {
    return countries
      .filter((c) => visitedCountryIds.has(c.id))
      .slice(0, 5);
  }, [countries, visitedCountryIds]);

  return (
    <ScrollView style={styles.container} showsVerticalScrollIndicator={false}>
      {/* World Map Visual */}
      <View style={styles.mapContainer}>
        <Text style={styles.mapTitle}>Explore the World</Text>

        {/* Region Grid */}
        <View style={styles.regionGrid}>
          {Object.entries(REGION_CONFIG).map(([regionKey, config]) => {
            const stats = regionStats[regionKey];
            const hasVisits = stats?.visited > 0;
            const percentage = stats?.total > 0
              ? Math.round((stats.visited / stats.total) * 100)
              : 0;

            return (
              <TouchableOpacity
                key={regionKey}
                style={[
                  styles.regionCard,
                  { backgroundColor: hasVisits ? config.visitedColor : config.color },
                ]}
                onPress={() => onRegionPress(regionKey)}
                activeOpacity={0.8}
              >
                <Text style={styles.regionEmoji}>{config.emoji}</Text>
                <Text style={[
                  styles.regionName,
                  hasVisits && styles.regionNameVisited,
                ]}>
                  {config.name}
                </Text>
                <Text style={[
                  styles.regionStats,
                  hasVisits && styles.regionStatsVisited,
                ]}>
                  {stats?.visited || 0}/{stats?.total || 0}
                </Text>
                {hasVisits && (
                  <View style={styles.progressBar}>
                    <View
                      style={[
                        styles.progressFill,
                        { width: `${percentage}%` },
                      ]}
                    />
                  </View>
                )}
              </TouchableOpacity>
            );
          })}
        </View>
      </View>

      {/* Recently Visited */}
      {recentlyVisited.length > 0 && (
        <View style={styles.recentSection}>
          <Text style={styles.sectionTitle}>Recently Visited</Text>
          <View style={styles.recentList}>
            {recentlyVisited.map((country) => (
              <TouchableOpacity
                key={country.id}
                style={styles.recentItem}
                onPress={() => onCountryPress(country)}
              >
                <Text style={styles.recentCode}>{country.isoCode}</Text>
                <Text style={styles.recentName} numberOfLines={1}>
                  {country.name}
                </Text>
              </TouchableOpacity>
            ))}
          </View>
        </View>
      )}

      {/* Legend */}
      <View style={styles.legend}>
        <View style={styles.legendItem}>
          <View style={[styles.legendDot, { backgroundColor: '#e5e7eb' }]} />
          <Text style={styles.legendText}>Not visited</Text>
        </View>
        <View style={styles.legendItem}>
          <View style={[styles.legendDot, { backgroundColor: '#22c55e' }]} />
          <Text style={styles.legendText}>Visited</Text>
        </View>
      </View>

      {/* Tap hint */}
      <Text style={styles.hint}>Tap a region to filter countries</Text>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  mapContainer: {
    backgroundColor: '#fff',
    borderRadius: 16,
    padding: 16,
    marginBottom: 16,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
  },
  mapTitle: {
    fontSize: 20,
    fontWeight: '700',
    color: '#111827',
    textAlign: 'center',
    marginBottom: 16,
  },
  regionGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    justifyContent: 'space-between',
    gap: 12,
  },
  regionCard: {
    width: '47%',
    borderRadius: 12,
    padding: 16,
    alignItems: 'center',
    minHeight: 120,
  },
  regionEmoji: {
    fontSize: 32,
    marginBottom: 8,
  },
  regionName: {
    fontSize: 14,
    fontWeight: '600',
    color: '#374151',
    marginBottom: 4,
  },
  regionNameVisited: {
    color: '#fff',
  },
  regionStats: {
    fontSize: 12,
    color: '#6b7280',
    marginBottom: 8,
  },
  regionStatsVisited: {
    color: 'rgba(255, 255, 255, 0.9)',
  },
  progressBar: {
    width: '100%',
    height: 4,
    backgroundColor: 'rgba(255, 255, 255, 0.3)',
    borderRadius: 2,
    overflow: 'hidden',
  },
  progressFill: {
    height: '100%',
    backgroundColor: '#fff',
    borderRadius: 2,
  },
  recentSection: {
    backgroundColor: '#fff',
    borderRadius: 16,
    padding: 16,
    marginBottom: 16,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
  },
  sectionTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: '#111827',
    marginBottom: 12,
  },
  recentList: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 8,
  },
  recentItem: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#f3f4f6',
    paddingHorizontal: 12,
    paddingVertical: 8,
    borderRadius: 20,
  },
  recentCode: {
    fontSize: 12,
    fontWeight: '700',
    color: '#4f46e5',
    marginRight: 6,
  },
  recentName: {
    fontSize: 14,
    color: '#374151',
    maxWidth: 100,
  },
  legend: {
    flexDirection: 'row',
    justifyContent: 'center',
    gap: 24,
    marginBottom: 8,
  },
  legendItem: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  legendDot: {
    width: 12,
    height: 12,
    borderRadius: 6,
    marginRight: 6,
  },
  legendText: {
    fontSize: 12,
    color: '#6b7280',
  },
  hint: {
    fontSize: 12,
    color: '#9ca3af',
    textAlign: 'center',
    marginBottom: 16,
  },
});

export default WorldMap;
