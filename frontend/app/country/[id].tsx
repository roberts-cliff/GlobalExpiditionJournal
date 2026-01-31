import React, { useState, useEffect, useCallback, useMemo } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TouchableOpacity,
  RefreshControl,
  Platform,
  Alert,
} from 'react-native';
import { useLocalSearchParams, useRouter, Stack } from 'expo-router';
import DateTimePicker from '@react-native-community/datetimepicker';
import {
  Loading,
  ErrorMessage,
  ScrapbookEntryCard,
  AddEntryModal,
} from '../../components';
import {
  Country,
  CountryListResponse,
  Visit,
  VisitListResponse,
  ScrapbookEntry,
  ScrapbookEntryListResponse,
  CreateScrapbookEntryRequest,
  CreateVisitRequest,
} from '../../types/api';
import { useGet, usePost } from '../../hooks/useApi';

export default function CountryDetailScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const router = useRouter();
  const countryId = parseInt(id || '0', 10);

  // State
  const [refreshing, setRefreshing] = useState(false);
  const [showAddModal, setShowAddModal] = useState(false);
  const [showDatePicker, setShowDatePicker] = useState(false);
  const [visitDate, setVisitDate] = useState(new Date());

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
    execute: fetchVisits,
  } = useGet<VisitListResponse>('/api/v1/visits');

  const {
    data: entriesData,
    loading: entriesLoading,
    execute: fetchEntries,
  } = useGet<ScrapbookEntryListResponse>('/api/v1/scrapbook/entries');

  const {
    execute: createVisit,
    loading: createVisitLoading,
  } = usePost<Visit, CreateVisitRequest>('/api/v1/visits');

  const {
    execute: createEntry,
    loading: createEntryLoading,
  } = usePost<ScrapbookEntry, CreateScrapbookEntryRequest>('/api/v1/scrapbook/entries');

  // Initial fetch
  useEffect(() => {
    fetchCountries();
    fetchVisits();
    fetchEntries();
  }, []);

  // Pull to refresh
  const onRefresh = useCallback(async () => {
    setRefreshing(true);
    await Promise.all([fetchCountries(), fetchVisits(), fetchEntries()]);
    setRefreshing(false);
  }, [fetchCountries, fetchVisits, fetchEntries]);

  // Get current country
  const country = useMemo(() => {
    return countriesData?.countries?.find((c) => c.id === countryId) || null;
  }, [countriesData, countryId]);

  // Get visit for this country
  const visit = useMemo(() => {
    return visitsData?.visits?.find((v) => v.countryId === countryId) || null;
  }, [visitsData, countryId]);

  // Get entries for this country
  const countryEntries = useMemo(() => {
    return entriesData?.entries?.filter((e) => e.countryId === countryId) || [];
  }, [entriesData, countryId]);

  // Handle mark as visited
  const handleMarkVisited = useCallback(async () => {
    if (Platform.OS === 'web') {
      // For web, just use current date
      const result = await createVisit({
        countryId,
        visitedAt: new Date().toISOString(),
      });
      if (result) {
        fetchVisits();
      } else {
        Alert.alert('Error', 'Failed to mark country as visited');
      }
    } else {
      setShowDatePicker(true);
    }
  }, [countryId, createVisit, fetchVisits]);

  // Handle date picker change
  const handleDateChange = useCallback(
    async (event: any, selectedDate?: Date) => {
      setShowDatePicker(Platform.OS === 'ios');
      if (selectedDate) {
        setVisitDate(selectedDate);
        if (Platform.OS !== 'ios') {
          // Android dismisses on selection
          const result = await createVisit({
            countryId,
            visitedAt: selectedDate.toISOString(),
          });
          if (result) {
            fetchVisits();
          } else {
            Alert.alert('Error', 'Failed to mark country as visited');
          }
        }
      }
    },
    [countryId, createVisit, fetchVisits]
  );

  // Handle iOS date confirm
  const handleDateConfirm = useCallback(async () => {
    setShowDatePicker(false);
    const result = await createVisit({
      countryId,
      visitedAt: visitDate.toISOString(),
    });
    if (result) {
      fetchVisits();
    } else {
      Alert.alert('Error', 'Failed to mark country as visited');
    }
  }, [countryId, visitDate, createVisit, fetchVisits]);

  // Handle save new entry
  const handleSaveEntry = useCallback(
    async (entryData: CreateScrapbookEntryRequest) => {
      const result = await createEntry(entryData);
      if (result) {
        setShowAddModal(false);
        fetchEntries();
      } else {
        Alert.alert('Error', 'Failed to create memory');
      }
    },
    [createEntry, fetchEntries]
  );

  // Handle entry press
  const handleEntryPress = useCallback((entry: ScrapbookEntry) => {
    router.push(`/scrapbook/${entry.id}`);
  }, [router]);

  // Format visit date
  const formattedVisitDate = useMemo(() => {
    if (!visit) return null;
    return new Date(visit.visitedAt).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  }, [visit]);

  // Loading state
  if (countriesLoading && !countriesData) {
    return <Loading message="Loading country..." />;
  }

  // Error state
  if (countriesError) {
    return <ErrorMessage message={countriesError} onRetry={fetchCountries} />;
  }

  // Country not found
  if (!country) {
    return (
      <View style={styles.centered}>
        <Text style={styles.errorText}>Country not found</Text>
        <TouchableOpacity style={styles.backButton} onPress={() => router.back()}>
          <Text style={styles.backButtonText}>Go Back</Text>
        </TouchableOpacity>
      </View>
    );
  }

  return (
    <>
      <Stack.Screen
        options={{
          title: country.name,
          headerStyle: { backgroundColor: '#4A90A4' },
          headerTintColor: '#fff',
        }}
      />

      <ScrollView
        style={styles.container}
        refreshControl={
          <RefreshControl refreshing={refreshing} onRefresh={onRefresh} />
        }
      >
        {/* Country Header */}
        <View style={styles.header}>
          <Text style={styles.countryName}>{country.name}</Text>
          <View style={styles.countryMeta}>
            <View style={styles.metaItem}>
              <Text style={styles.metaLabel}>Region</Text>
              <Text style={styles.metaValue}>{country.region}</Text>
            </View>
            <View style={styles.metaDivider} />
            <View style={styles.metaItem}>
              <Text style={styles.metaLabel}>ISO Code</Text>
              <Text style={styles.metaValue}>{country.isoCode}</Text>
            </View>
          </View>
        </View>

        {/* Visit Status */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Visit Status</Text>
          <View style={styles.visitCard}>
            {visit ? (
              <>
                <View style={styles.visitedBadge}>
                  <Text style={styles.visitedText}>Visited</Text>
                </View>
                <Text style={styles.visitDate}>{formattedVisitDate}</Text>
              </>
            ) : (
              <>
                <Text style={styles.notVisitedText}>Not visited yet</Text>
                <TouchableOpacity
                  style={styles.markVisitedButton}
                  onPress={handleMarkVisited}
                  disabled={createVisitLoading}
                >
                  <Text style={styles.markVisitedButtonText}>
                    {createVisitLoading ? 'Saving...' : 'Mark as Visited'}
                  </Text>
                </TouchableOpacity>
              </>
            )}
          </View>
        </View>

        {/* Date Picker for iOS */}
        {showDatePicker && Platform.OS === 'ios' && (
          <View style={styles.datePickerContainer}>
            <DateTimePicker
              value={visitDate}
              mode="date"
              display="spinner"
              onChange={handleDateChange}
              maximumDate={new Date()}
            />
            <View style={styles.datePickerButtons}>
              <TouchableOpacity
                style={styles.datePickerCancel}
                onPress={() => setShowDatePicker(false)}
              >
                <Text style={styles.datePickerCancelText}>Cancel</Text>
              </TouchableOpacity>
              <TouchableOpacity
                style={styles.datePickerConfirm}
                onPress={handleDateConfirm}
              >
                <Text style={styles.datePickerConfirmText}>Confirm</Text>
              </TouchableOpacity>
            </View>
          </View>
        )}

        {/* Date Picker for Android */}
        {showDatePicker && Platform.OS === 'android' && (
          <DateTimePicker
            value={visitDate}
            mode="date"
            display="default"
            onChange={handleDateChange}
            maximumDate={new Date()}
          />
        )}

        {/* Scrapbook Entries */}
        <View style={styles.section}>
          <View style={styles.sectionHeader}>
            <Text style={styles.sectionTitle}>Memories</Text>
            <TouchableOpacity
              style={styles.addMemoryButton}
              onPress={() => setShowAddModal(true)}
            >
              <Text style={styles.addMemoryButtonText}>+ Add</Text>
            </TouchableOpacity>
          </View>

          {entriesLoading && !entriesData ? (
            <View style={styles.loadingContainer}>
              <Text style={styles.loadingText}>Loading memories...</Text>
            </View>
          ) : countryEntries.length > 0 ? (
            <View style={styles.entriesList}>
              {countryEntries.map((entry) => (
                <ScrapbookEntryCard
                  key={entry.id}
                  entry={entry}
                  onPress={handleEntryPress}
                />
              ))}
            </View>
          ) : (
            <View style={styles.emptyState}>
              <Text style={styles.emptyIcon}>📸</Text>
              <Text style={styles.emptyTitle}>No memories yet</Text>
              <Text style={styles.emptyText}>
                Add your first memory for {country.name}!
              </Text>
              <TouchableOpacity
                style={styles.emptyButton}
                onPress={() => setShowAddModal(true)}
              >
                <Text style={styles.emptyButtonText}>Add Memory</Text>
              </TouchableOpacity>
            </View>
          )}
        </View>
      </ScrollView>

      <AddEntryModal
        visible={showAddModal}
        onClose={() => setShowAddModal(false)}
        onSave={handleSaveEntry}
        countries={country ? [country] : []}
        loading={createEntryLoading}
      />
    </>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f5f5f5',
  },
  centered: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#f5f5f5',
  },
  errorText: {
    fontSize: 18,
    color: '#6b7280',
    marginBottom: 16,
  },
  backButton: {
    backgroundColor: '#4f46e5',
    paddingHorizontal: 24,
    paddingVertical: 12,
    borderRadius: 8,
  },
  backButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
  header: {
    backgroundColor: '#fff',
    padding: 20,
    marginBottom: 16,
  },
  countryName: {
    fontSize: 32,
    fontWeight: 'bold',
    color: '#111827',
    marginBottom: 16,
  },
  countryMeta: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  metaItem: {
    flex: 1,
  },
  metaLabel: {
    fontSize: 12,
    color: '#6b7280',
    marginBottom: 4,
  },
  metaValue: {
    fontSize: 16,
    fontWeight: '600',
    color: '#374151',
  },
  metaDivider: {
    width: 1,
    height: 40,
    backgroundColor: '#e5e7eb',
    marginHorizontal: 16,
  },
  section: {
    backgroundColor: '#fff',
    padding: 16,
    marginBottom: 16,
  },
  sectionHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 16,
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#111827',
    marginBottom: 12,
  },
  visitCard: {
    backgroundColor: '#f9fafb',
    borderRadius: 12,
    padding: 16,
    alignItems: 'center',
  },
  visitedBadge: {
    backgroundColor: '#22c55e',
    paddingHorizontal: 16,
    paddingVertical: 6,
    borderRadius: 20,
    marginBottom: 8,
  },
  visitedText: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '600',
  },
  visitDate: {
    fontSize: 16,
    color: '#374151',
  },
  notVisitedText: {
    fontSize: 16,
    color: '#6b7280',
    marginBottom: 12,
  },
  markVisitedButton: {
    backgroundColor: '#4f46e5',
    paddingHorizontal: 24,
    paddingVertical: 12,
    borderRadius: 8,
  },
  markVisitedButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
  datePickerContainer: {
    backgroundColor: '#fff',
    padding: 16,
    marginBottom: 16,
  },
  datePickerButtons: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginTop: 16,
  },
  datePickerCancel: {
    paddingHorizontal: 24,
    paddingVertical: 12,
  },
  datePickerCancelText: {
    fontSize: 16,
    color: '#6b7280',
  },
  datePickerConfirm: {
    backgroundColor: '#4f46e5',
    paddingHorizontal: 24,
    paddingVertical: 12,
    borderRadius: 8,
  },
  datePickerConfirmText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
  addMemoryButton: {
    backgroundColor: '#4f46e5',
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 20,
  },
  addMemoryButtonText: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '600',
  },
  loadingContainer: {
    padding: 24,
    alignItems: 'center',
  },
  loadingText: {
    fontSize: 14,
    color: '#6b7280',
  },
  entriesList: {
    marginTop: 8,
  },
  emptyState: {
    alignItems: 'center',
    paddingVertical: 32,
  },
  emptyIcon: {
    fontSize: 48,
    marginBottom: 12,
  },
  emptyTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: '#374151',
    marginBottom: 4,
  },
  emptyText: {
    fontSize: 14,
    color: '#6b7280',
    textAlign: 'center',
    marginBottom: 16,
  },
  emptyButton: {
    backgroundColor: '#4f46e5',
    paddingHorizontal: 20,
    paddingVertical: 10,
    borderRadius: 8,
  },
  emptyButtonText: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '600',
  },
});
