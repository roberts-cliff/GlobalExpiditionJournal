import React, { useState, useEffect, useCallback, useMemo } from 'react';
import {
  View,
  Text,
  StyleSheet,
  FlatList,
  RefreshControl,
  TouchableOpacity,
  Alert,
} from 'react-native';
import { useRouter } from 'expo-router';
import {
  SearchBar,
  FilterChips,
  Loading,
  ErrorMessage,
  ScrapbookEntryCard,
  AddEntryModal,
} from '../../components';
import {
  CountryListResponse,
  ScrapbookEntry,
  ScrapbookEntryListResponse,
  CreateScrapbookEntryRequest,
} from '../../types/api';
import { useGet, usePost } from '../../hooks/useApi';

export default function ScrapbookScreen() {
  const router = useRouter();

  // State
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedCountry, setSelectedCountry] = useState<string | null>(null);
  const [selectedTag, setSelectedTag] = useState<string | null>(null);
  const [refreshing, setRefreshing] = useState(false);
  const [showAddModal, setShowAddModal] = useState(false);

  // API calls
  const {
    data: entriesData,
    loading: entriesLoading,
    error: entriesError,
    execute: fetchEntries,
  } = useGet<ScrapbookEntryListResponse>('/api/v1/scrapbook/entries');

  const {
    data: countriesData,
    execute: fetchCountries,
  } = useGet<CountryListResponse>('/api/v1/countries');

  const {
    execute: createEntry,
    loading: createLoading,
  } = usePost<ScrapbookEntry, CreateScrapbookEntryRequest>('/api/v1/scrapbook/entries');

  // Initial fetch
  useEffect(() => {
    fetchEntries();
    fetchCountries();
  }, []);

  // Pull to refresh
  const onRefresh = useCallback(async () => {
    setRefreshing(true);
    await Promise.all([fetchEntries(), fetchCountries()]);
    setRefreshing(false);
  }, [fetchEntries, fetchCountries]);

  // Get unique countries from entries for filtering
  const countryFilters = useMemo(() => {
    if (!entriesData?.entries) return [];
    const countries = new Set<string>();
    entriesData.entries.forEach((entry) => {
      if (entry.country?.name) {
        countries.add(entry.country.name);
      }
    });
    return Array.from(countries).sort();
  }, [entriesData]);

  // Get unique tags from entries for filtering
  const tagFilters = useMemo(() => {
    if (!entriesData?.entries) return [];
    const tags = new Set<string>();
    entriesData.entries.forEach((entry) => {
      if (entry.tags) {
        entry.tags.split(',').forEach((tag) => {
          const trimmed = tag.trim();
          if (trimmed) tags.add(trimmed);
        });
      }
    });
    return Array.from(tags).sort();
  }, [entriesData]);

  // Filter entries
  const filteredEntries = useMemo(() => {
    if (!entriesData?.entries) return [];

    return entriesData.entries.filter((entry) => {
      // Search filter (now includes tags)
      if (searchQuery) {
        const query = searchQuery.toLowerCase();
        if (
          !entry.title.toLowerCase().includes(query) &&
          !(entry.notes?.toLowerCase().includes(query)) &&
          !(entry.country?.name.toLowerCase().includes(query)) &&
          !(entry.tags?.toLowerCase().includes(query))
        ) {
          return false;
        }
      }

      // Country filter
      if (selectedCountry && entry.country?.name !== selectedCountry) {
        return false;
      }

      // Tag filter
      if (selectedTag && !entry.tags?.split(',').map(t => t.trim()).includes(selectedTag)) {
        return false;
      }

      return true;
    });
  }, [entriesData, searchQuery, selectedCountry, selectedTag]);

  // Stats
  const stats = useMemo(() => {
    const totalEntries = entriesData?.entries?.length || 0;
    const countriesDocumented = new Set(
      entriesData?.entries?.map((e) => e.countryId) || []
    ).size;
    const photosUploaded = entriesData?.entries?.filter((e) => e.mediaUrl)?.length || 0;
    return { totalEntries, countriesDocumented, photosUploaded };
  }, [entriesData]);

  // Handle entry press
  const handleEntryPress = useCallback((entry: ScrapbookEntry) => {
    router.push(`/scrapbook/${entry.id}`);
  }, [router]);

  // Handle save new entry
  const handleSaveEntry = useCallback(async (entryData: CreateScrapbookEntryRequest) => {
    const result = await createEntry(entryData);
    if (result) {
      setShowAddModal(false);
      fetchEntries();
    } else {
      Alert.alert('Error', 'Failed to create entry. Please try again.');
    }
  }, [createEntry, fetchEntries]);

  // Render entry item
  const renderEntryItem = useCallback(
    ({ item }: { item: ScrapbookEntry }) => (
      <ScrapbookEntryCard entry={item} onPress={handleEntryPress} />
    ),
    [handleEntryPress]
  );

  // Loading state
  if (entriesLoading && !entriesData) {
    return <Loading message="Loading your memories..." />;
  }

  // Error state
  if (entriesError) {
    return (
      <ErrorMessage
        message={entriesError}
        onRetry={() => {
          fetchEntries();
          fetchCountries();
        }}
      />
    );
  }

  return (
    <View style={styles.container}>
      <FlatList
        data={filteredEntries}
        keyExtractor={(item) => item.id.toString()}
        renderItem={renderEntryItem}
        numColumns={1}
        refreshControl={
          <RefreshControl refreshing={refreshing} onRefresh={onRefresh} />
        }
        ListHeaderComponent={
          <View style={styles.header}>
            <View style={styles.titleRow}>
              <Text style={styles.title}>My Scrapbook</Text>
              <TouchableOpacity
                style={styles.addButton}
                onPress={() => setShowAddModal(true)}
              >
                <Text style={styles.addButtonText}>+ Add</Text>
              </TouchableOpacity>
            </View>

            {/* Stats Row */}
            <View style={styles.statsRow}>
              <View style={styles.statItem}>
                <Text style={styles.statNumber}>{stats.totalEntries}</Text>
                <Text style={styles.statLabel}>Memories</Text>
              </View>
              <View style={styles.statDivider} />
              <View style={styles.statItem}>
                <Text style={styles.statNumber}>{stats.countriesDocumented}</Text>
                <Text style={styles.statLabel}>Countries</Text>
              </View>
              <View style={styles.statDivider} />
              <View style={styles.statItem}>
                <Text style={styles.statNumber}>{stats.photosUploaded}</Text>
                <Text style={styles.statLabel}>Photos</Text>
              </View>
            </View>

            <SearchBar
              value={searchQuery}
              onChangeText={setSearchQuery}
              placeholder="Search memories..."
            />

            {countryFilters.length > 0 && (
              <FilterChips
                options={countryFilters}
                selected={selectedCountry}
                onSelect={setSelectedCountry}
                label="Filter by country"
              />
            )}

            {tagFilters.length > 0 && (
              <FilterChips
                options={tagFilters}
                selected={selectedTag}
                onSelect={setSelectedTag}
                label="Filter by tag"
              />
            )}

            <Text style={styles.countText}>
              {filteredEntries.length} {filteredEntries.length === 1 ? 'memory' : 'memories'}
            </Text>
          </View>
        }
        ListEmptyComponent={
          <View style={styles.emptyState}>
            <Text style={styles.emptyIcon}>📸</Text>
            <Text style={styles.emptyTitle}>
              {searchQuery || selectedCountry || selectedTag
                ? 'No memories found'
                : 'Start your scrapbook'}
            </Text>
            <Text style={styles.emptyText}>
              {searchQuery || selectedCountry || selectedTag
                ? 'Try adjusting your search or filters'
                : 'Capture your travel memories by adding your first entry!'}
            </Text>
            {!searchQuery && !selectedCountry && !selectedTag && (
              <TouchableOpacity
                style={styles.emptyButton}
                onPress={() => setShowAddModal(true)}
              >
                <Text style={styles.emptyButtonText}>Add First Memory</Text>
              </TouchableOpacity>
            )}
          </View>
        }
        contentContainerStyle={styles.listContent}
      />

      <AddEntryModal
        visible={showAddModal}
        onClose={() => setShowAddModal(false)}
        onSave={handleSaveEntry}
        countries={countriesData?.countries || []}
        loading={createLoading}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f5f5f5',
  },
  listContent: {
    padding: 16,
    paddingBottom: 32,
  },
  header: {
    marginBottom: 8,
  },
  titleRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 16,
  },
  title: {
    fontSize: 28,
    fontWeight: 'bold',
    color: '#111827',
  },
  addButton: {
    backgroundColor: '#4f46e5',
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 20,
  },
  addButtonText: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '600',
  },
  statsRow: {
    flexDirection: 'row',
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 16,
    marginBottom: 16,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
  },
  statItem: {
    flex: 1,
    alignItems: 'center',
  },
  statNumber: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#4f46e5',
    marginBottom: 4,
  },
  statLabel: {
    fontSize: 12,
    color: '#6b7280',
  },
  statDivider: {
    width: 1,
    backgroundColor: '#e5e7eb',
    marginVertical: 4,
  },
  countText: {
    fontSize: 14,
    color: '#6b7280',
    marginBottom: 8,
  },
  emptyState: {
    alignItems: 'center',
    paddingVertical: 48,
  },
  emptyIcon: {
    fontSize: 64,
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
    marginBottom: 24,
  },
  emptyButton: {
    backgroundColor: '#4f46e5',
    paddingHorizontal: 24,
    paddingVertical: 12,
    borderRadius: 8,
  },
  emptyButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
});
