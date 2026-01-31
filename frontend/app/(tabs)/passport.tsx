import React, { useState, useEffect, useCallback, useMemo } from 'react';
import {
  View,
  Text,
  StyleSheet,
  FlatList,
  RefreshControl,
  TouchableOpacity,
} from 'react-native';
import { useRouter } from 'expo-router';
import { useAuth } from '../../context/AuthContext';
import {
  StatsCard,
  SearchBar,
  FilterChips,
  CountryListItem,
  Loading,
  ErrorMessage,
  WorldMap,
} from '../../components';
import { Country, Visit, CountryListResponse, VisitListResponse } from '../../types/api';
import { useGet } from '../../hooks/useApi';

// Regions for filtering
const REGIONS = ['Africa', 'Americas', 'Asia', 'Europe', 'Oceania'];

export default function PassportScreen() {
  const { user } = useAuth();
  const router = useRouter();

  // State
  const [viewMode, setViewMode] = useState<'list' | 'map'>('list');
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedRegion, setSelectedRegion] = useState<string | null>(null);
  const [showVisitedOnly, setShowVisitedOnly] = useState(false);
  const [refreshing, setRefreshing] = useState(false);

  // API calls
  const {
    data: countriesData,
    loading: countriesLoading,
    error: countriesError,
    execute: fetchCountries,
  } = useGet<CountryListResponse>('/api/v1/countries');

  const {
    data: visitsData,
    loading: visitsLoading,
    error: visitsError,
    execute: fetchVisits,
  } = useGet<VisitListResponse>('/api/v1/visits');

  // Initial fetch
  useEffect(() => {
    fetchCountries();
    fetchVisits();
  }, []);

  // Pull to refresh
  const onRefresh = useCallback(async () => {
    setRefreshing(true);
    await Promise.all([fetchCountries(), fetchVisits()]);
    setRefreshing(false);
  }, [fetchCountries, fetchVisits]);

  // Create a map of visited countries for quick lookup
  const visitedCountryMap = useMemo(() => {
    const map = new Map<number, Visit>();
    if (visitsData?.visits) {
      visitsData.visits.forEach((visit) => {
        // Keep the most recent visit for each country
        const existing = map.get(visit.countryId);
        if (!existing || new Date(visit.visitedAt) > new Date(existing.visitedAt)) {
          map.set(visit.countryId, visit);
        }
      });
    }
    return map;
  }, [visitsData]);

  // Filter countries
  const filteredCountries = useMemo(() => {
    if (!countriesData?.countries) return [];

    return countriesData.countries.filter((country) => {
      // Search filter
      if (searchQuery) {
        const query = searchQuery.toLowerCase();
        if (
          !country.name.toLowerCase().includes(query) &&
          !country.isoCode.toLowerCase().includes(query)
        ) {
          return false;
        }
      }

      // Region filter
      if (selectedRegion && country.region !== selectedRegion) {
        return false;
      }

      // Visited filter
      if (showVisitedOnly && !visitedCountryMap.has(country.id)) {
        return false;
      }

      return true;
    });
  }, [countriesData, searchQuery, selectedRegion, showVisitedOnly, visitedCountryMap]);

  // Stats
  const stats = useMemo(() => {
    const totalCountries = countriesData?.countries?.length || 0;
    const visitedCount = visitedCountryMap.size;
    return { totalCountries, visitedCount };
  }, [countriesData, visitedCountryMap]);

  // Visited country IDs set for WorldMap
  const visitedCountryIds = useMemo(() => {
    return new Set(visitedCountryMap.keys());
  }, [visitedCountryMap]);

  // Handle region press from map (sets filter and switches to list)
  const handleRegionPress = useCallback((region: string) => {
    setSelectedRegion(region);
    setViewMode('list');
  }, []);

  // Handle country press
  const handleCountryPress = useCallback((country: Country) => {
    router.push(`/country/${country.id}`);
  }, [router]);

  // Render country item
  const renderCountryItem = useCallback(
    ({ item }: { item: Country }) => {
      const visit = visitedCountryMap.get(item.id);
      return (
        <CountryListItem
          country={item}
          isVisited={!!visit}
          visitDate={visit?.visitedAt}
          onPress={handleCountryPress}
        />
      );
    },
    [visitedCountryMap, handleCountryPress]
  );

  // Loading state
  if (countriesLoading && !countriesData) {
    return <Loading message="Loading countries..." />;
  }

  // Error state
  if (countriesError) {
    return (
      <ErrorMessage
        message={countriesError}
        onRetry={() => {
          fetchCountries();
          fetchVisits();
        }}
      />
    );
  }

  return (
    <View style={styles.container}>
      {/* Header - always visible */}
      <View style={styles.headerFixed}>
        <View style={styles.titleRow}>
          <View>
            <Text style={styles.title}>My Passport</Text>
            <Text style={styles.welcome}>
              Welcome, {user?.displayName || 'Explorer'}!
            </Text>
          </View>
          {/* View Mode Toggle */}
          <View style={styles.viewToggle}>
            <TouchableOpacity
              style={[styles.viewToggleBtn, viewMode === 'map' && styles.viewToggleBtnActive]}
              onPress={() => setViewMode('map')}
            >
              <Text style={[styles.viewToggleText, viewMode === 'map' && styles.viewToggleTextActive]}>
                🗺️ Map
              </Text>
            </TouchableOpacity>
            <TouchableOpacity
              style={[styles.viewToggleBtn, viewMode === 'list' && styles.viewToggleBtnActive]}
              onPress={() => setViewMode('list')}
            >
              <Text style={[styles.viewToggleText, viewMode === 'list' && styles.viewToggleTextActive]}>
                📋 List
              </Text>
            </TouchableOpacity>
          </View>
        </View>

        <StatsCard
          countriesVisited={stats.visitedCount}
          totalCountries={stats.totalCountries}
        />
      </View>

      {/* Conditional View */}
      {viewMode === 'map' ? (
        <View style={styles.mapContainer}>
          <WorldMap
            countries={countriesData?.countries || []}
            visitedCountryIds={visitedCountryIds}
            onRegionPress={handleRegionPress}
            onCountryPress={handleCountryPress}
          />
        </View>
      ) : (
        <FlatList
          data={filteredCountries}
          keyExtractor={(item) => item.id.toString()}
          renderItem={renderCountryItem}
          refreshControl={
            <RefreshControl refreshing={refreshing} onRefresh={onRefresh} />
          }
          ListHeaderComponent={
            <View style={styles.listHeader}>
              <SearchBar
                value={searchQuery}
                onChangeText={setSearchQuery}
                placeholder="Search countries..."
              />

              <FilterChips
                options={REGIONS}
                selected={selectedRegion}
                onSelect={setSelectedRegion}
                label="Filter by region"
              />

              <View style={styles.toggleRow}>
                <TouchableOpacity
                  style={[
                    styles.toggleButton,
                    showVisitedOnly && styles.toggleButtonActive,
                  ]}
                  onPress={() => setShowVisitedOnly(!showVisitedOnly)}
                >
                  <Text
                    style={[
                      styles.toggleText,
                      showVisitedOnly && styles.toggleTextActive,
                    ]}
                  >
                    {showVisitedOnly ? '✓ Visited only' : 'Show all'}
                  </Text>
                </TouchableOpacity>

                <Text style={styles.countText}>
                  {filteredCountries.length} countries
                </Text>
              </View>
            </View>
          }
          ListEmptyComponent={
            <View style={styles.emptyState}>
              <Text style={styles.emptyIcon}>🌍</Text>
              <Text style={styles.emptyTitle}>
                {showVisitedOnly
                  ? 'No visited countries yet'
                  : 'No countries found'}
              </Text>
              <Text style={styles.emptyText}>
                {showVisitedOnly
                  ? 'Start your journey by exploring countries and marking them as visited!'
                  : 'Try adjusting your search or filters'}
              </Text>
            </View>
          }
          contentContainerStyle={styles.listContent}
        />
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f5f5f5',
  },
  headerFixed: {
    backgroundColor: '#f5f5f5',
    padding: 16,
    paddingBottom: 8,
  },
  titleRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    marginBottom: 16,
  },
  viewToggle: {
    flexDirection: 'row',
    backgroundColor: '#e5e7eb',
    borderRadius: 8,
    padding: 2,
  },
  viewToggleBtn: {
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 6,
  },
  viewToggleBtnActive: {
    backgroundColor: '#fff',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.1,
    shadowRadius: 2,
    elevation: 1,
  },
  viewToggleText: {
    fontSize: 13,
    color: '#6b7280',
  },
  viewToggleTextActive: {
    color: '#111827',
    fontWeight: '600',
  },
  mapContainer: {
    flex: 1,
    padding: 16,
    paddingTop: 0,
  },
  listContent: {
    padding: 16,
    paddingTop: 0,
    paddingBottom: 32,
  },
  listHeader: {
    marginBottom: 8,
  },
  title: {
    fontSize: 28,
    fontWeight: 'bold',
    color: '#111827',
    marginBottom: 4,
  },
  welcome: {
    fontSize: 16,
    color: '#6b7280',
    marginBottom: 0,
  },
  toggleRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 16,
  },
  toggleButton: {
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 20,
    backgroundColor: '#f3f4f6',
  },
  toggleButtonActive: {
    backgroundColor: '#22c55e',
  },
  toggleText: {
    fontSize: 14,
    color: '#374151',
  },
  toggleTextActive: {
    color: '#fff',
    fontWeight: '600',
  },
  countText: {
    fontSize: 14,
    color: '#6b7280',
  },
  emptyState: {
    alignItems: 'center',
    paddingVertical: 48,
  },
  emptyIcon: {
    fontSize: 48,
    marginBottom: 16,
  },
  emptyTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#374151',
    marginBottom: 8,
  },
  emptyText: {
    fontSize: 14,
    color: '#6b7280',
    textAlign: 'center',
    paddingHorizontal: 32,
  },
});
