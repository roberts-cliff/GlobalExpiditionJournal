import React, { useState, useEffect, useCallback, useMemo } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TouchableOpacity,
  RefreshControl,
  Image,
  TextInput,
  Alert,
  ActivityIndicator,
} from 'react-native';
import { useLocalSearchParams, useRouter, Stack } from 'expo-router';
import { Loading, ErrorMessage } from '../../components';
import { ScrapbookEntry, UpdateScrapbookEntryRequest } from '../../types/api';
import { useGet, usePut, useDelete } from '../../hooks/useApi';

export default function ScrapbookEntryDetailScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const router = useRouter();
  const entryId = parseInt(id || '0', 10);

  // State
  const [refreshing, setRefreshing] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [editTitle, setEditTitle] = useState('');
  const [editNotes, setEditNotes] = useState('');

  // API calls
  const {
    data: entry,
    loading: entryLoading,
    error: entryError,
    execute: fetchEntry,
  } = useGet<ScrapbookEntry>(`/api/v1/scrapbook/entries/${entryId}`);

  const {
    execute: updateEntry,
    loading: updateLoading,
  } = usePut<ScrapbookEntry, UpdateScrapbookEntryRequest>(`/api/v1/scrapbook/entries/${entryId}`);

  const {
    execute: deleteEntry,
    loading: deleteLoading,
  } = useDelete<{ message: string }>(`/api/v1/scrapbook/entries/${entryId}`);

  // Initial fetch
  useEffect(() => {
    if (entryId > 0) {
      fetchEntry();
    }
  }, [entryId]);

  // Initialize edit form when entry loads
  useEffect(() => {
    if (entry) {
      setEditTitle(entry.title);
      setEditNotes(entry.notes || '');
    }
  }, [entry]);

  // Pull to refresh
  const onRefresh = useCallback(async () => {
    setRefreshing(true);
    await fetchEntry();
    setRefreshing(false);
  }, [fetchEntry]);

  // Handle edit toggle
  const handleEditToggle = useCallback(() => {
    if (isEditing) {
      // Cancel editing - restore original values
      setEditTitle(entry?.title || '');
      setEditNotes(entry?.notes || '');
    }
    setIsEditing(!isEditing);
  }, [isEditing, entry]);

  // Handle save
  const handleSave = useCallback(async () => {
    if (!editTitle.trim()) {
      Alert.alert('Error', 'Title is required');
      return;
    }

    const result = await updateEntry({
      title: editTitle.trim(),
      notes: editNotes.trim() || undefined,
    });

    if (result) {
      setIsEditing(false);
      fetchEntry();
    } else {
      Alert.alert('Error', 'Failed to update memory. Please try again.');
    }
  }, [editTitle, editNotes, updateEntry, fetchEntry]);

  // Handle delete
  const handleDelete = useCallback(() => {
    Alert.alert(
      'Delete Memory',
      'Are you sure you want to delete this memory? This action cannot be undone.',
      [
        { text: 'Cancel', style: 'cancel' },
        {
          text: 'Delete',
          style: 'destructive',
          onPress: async () => {
            const result = await deleteEntry();
            if (result) {
              router.back();
            } else {
              Alert.alert('Error', 'Failed to delete memory. Please try again.');
            }
          },
        },
      ]
    );
  }, [deleteEntry, router]);

  // Format date
  const formattedDate = useMemo(() => {
    if (!entry?.createdAt) return '';
    return new Date(entry.createdAt).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  }, [entry]);

  const formattedVisitDate = useMemo(() => {
    if (!entry?.visitedAt) return null;
    return new Date(entry.visitedAt).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  }, [entry]);

  // Loading state
  if (entryLoading && !entry) {
    return <Loading message="Loading memory..." />;
  }

  // Error state
  if (entryError) {
    return <ErrorMessage message={entryError} onRetry={fetchEntry} />;
  }

  // Entry not found
  if (!entry) {
    return (
      <View style={styles.centered}>
        <Text style={styles.errorText}>Memory not found</Text>
        <TouchableOpacity style={styles.backButton} onPress={() => router.back()}>
          <Text style={styles.backButtonText}>Go Back</Text>
        </TouchableOpacity>
      </View>
    );
  }

  const isBusy = updateLoading || deleteLoading;

  return (
    <>
      <Stack.Screen
        options={{
          title: isEditing ? 'Edit Memory' : 'Memory',
          headerStyle: { backgroundColor: '#4A90A4' },
          headerTintColor: '#fff',
          headerRight: () => (
            <View style={styles.headerButtons}>
              {isEditing ? (
                <>
                  <TouchableOpacity
                    style={styles.headerButton}
                    onPress={handleEditToggle}
                    disabled={isBusy}
                  >
                    <Text style={styles.headerButtonText}>Cancel</Text>
                  </TouchableOpacity>
                  <TouchableOpacity
                    style={[styles.headerButton, styles.saveButton]}
                    onPress={handleSave}
                    disabled={isBusy || !editTitle.trim()}
                  >
                    {updateLoading ? (
                      <ActivityIndicator size="small" color="#fff" />
                    ) : (
                      <Text style={styles.saveButtonText}>Save</Text>
                    )}
                  </TouchableOpacity>
                </>
              ) : (
                <>
                  <TouchableOpacity
                    style={styles.headerButton}
                    onPress={handleEditToggle}
                  >
                    <Text style={styles.headerButtonText}>Edit</Text>
                  </TouchableOpacity>
                  <TouchableOpacity
                    style={[styles.headerButton, styles.deleteButton]}
                    onPress={handleDelete}
                    disabled={isBusy}
                  >
                    {deleteLoading ? (
                      <ActivityIndicator size="small" color="#fff" />
                    ) : (
                      <Text style={styles.deleteButtonText}>Delete</Text>
                    )}
                  </TouchableOpacity>
                </>
              )}
            </View>
          ),
        }}
      />

      <ScrollView
        style={styles.container}
        refreshControl={
          <RefreshControl refreshing={refreshing} onRefresh={onRefresh} />
        }
      >
        {/* Photo */}
        {entry.mediaUrl ? (
          <Image source={{ uri: entry.mediaUrl }} style={styles.photo} />
        ) : (
          <View style={styles.photoPlaceholder}>
            <Text style={styles.photoPlaceholderIcon}>📷</Text>
            <Text style={styles.photoPlaceholderText}>No photo</Text>
          </View>
        )}

        {/* Content */}
        <View style={styles.content}>
          {/* Country Badge */}
          {entry.country && (
            <TouchableOpacity
              style={styles.countryBadge}
              onPress={() => router.push(`/country/${entry.countryId}`)}
            >
              <Text style={styles.countryBadgeCode}>{entry.country.isoCode}</Text>
              <Text style={styles.countryBadgeName}>{entry.country.name}</Text>
            </TouchableOpacity>
          )}

          {/* Title */}
          {isEditing ? (
            <TextInput
              style={styles.titleInput}
              value={editTitle}
              onChangeText={setEditTitle}
              placeholder="Title"
              placeholderTextColor="#9ca3af"
              maxLength={100}
            />
          ) : (
            <Text style={styles.title}>{entry.title}</Text>
          )}

          {/* Dates */}
          <View style={styles.dateRow}>
            <Text style={styles.dateLabel}>Created:</Text>
            <Text style={styles.dateValue}>{formattedDate}</Text>
          </View>
          {formattedVisitDate && (
            <View style={styles.dateRow}>
              <Text style={styles.dateLabel}>Visited:</Text>
              <Text style={styles.dateValue}>{formattedVisitDate}</Text>
            </View>
          )}

          {/* Tags */}
          {entry.tags && (
            <View style={styles.tagsContainer}>
              {entry.tags.split(',').map((tag, index) => (
                <View key={index} style={styles.tag}>
                  <Text style={styles.tagText}>{tag.trim()}</Text>
                </View>
              ))}
            </View>
          )}

          {/* Notes */}
          <View style={styles.notesSection}>
            <Text style={styles.notesLabel}>Notes</Text>
            {isEditing ? (
              <TextInput
                style={styles.notesInput}
                value={editNotes}
                onChangeText={setEditNotes}
                placeholder="Write about this memory..."
                placeholderTextColor="#9ca3af"
                multiline
                numberOfLines={6}
                textAlignVertical="top"
              />
            ) : (
              <Text style={styles.notesText}>
                {entry.notes || 'No notes added.'}
              </Text>
            )}
          </View>
        </View>
      </ScrollView>
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
  headerButtons: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
  },
  headerButton: {
    paddingHorizontal: 12,
    paddingVertical: 6,
  },
  headerButtonText: {
    color: '#fff',
    fontSize: 16,
  },
  saveButton: {
    backgroundColor: '#22c55e',
    borderRadius: 6,
  },
  saveButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
  deleteButton: {
    backgroundColor: '#ef4444',
    borderRadius: 6,
  },
  deleteButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
  photo: {
    width: '100%',
    height: 300,
    backgroundColor: '#e5e7eb',
  },
  photoPlaceholder: {
    width: '100%',
    height: 200,
    backgroundColor: '#e5e7eb',
    justifyContent: 'center',
    alignItems: 'center',
  },
  photoPlaceholderIcon: {
    fontSize: 48,
    marginBottom: 8,
  },
  photoPlaceholderText: {
    fontSize: 16,
    color: '#9ca3af',
  },
  content: {
    padding: 20,
    backgroundColor: '#fff',
    minHeight: 300,
  },
  countryBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#eef2ff',
    alignSelf: 'flex-start',
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 20,
    marginBottom: 16,
  },
  countryBadgeCode: {
    fontSize: 12,
    fontWeight: '700',
    color: '#4f46e5',
    marginRight: 6,
  },
  countryBadgeName: {
    fontSize: 14,
    color: '#4f46e5',
  },
  title: {
    fontSize: 28,
    fontWeight: 'bold',
    color: '#111827',
    marginBottom: 12,
  },
  titleInput: {
    fontSize: 28,
    fontWeight: 'bold',
    color: '#111827',
    marginBottom: 12,
    borderBottomWidth: 2,
    borderBottomColor: '#4f46e5',
    paddingBottom: 8,
  },
  dateRow: {
    flexDirection: 'row',
    marginBottom: 4,
  },
  dateLabel: {
    fontSize: 14,
    color: '#6b7280',
    marginRight: 8,
  },
  dateValue: {
    fontSize: 14,
    color: '#374151',
  },
  tagsContainer: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    marginTop: 12,
    gap: 8,
  },
  tag: {
    backgroundColor: '#f3f4f6',
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 12,
  },
  tagText: {
    fontSize: 12,
    color: '#6b7280',
  },
  notesSection: {
    marginTop: 24,
  },
  notesLabel: {
    fontSize: 16,
    fontWeight: '600',
    color: '#374151',
    marginBottom: 8,
  },
  notesText: {
    fontSize: 16,
    color: '#4b5563',
    lineHeight: 24,
  },
  notesInput: {
    fontSize: 16,
    color: '#111827',
    lineHeight: 24,
    backgroundColor: '#f9fafb',
    borderRadius: 8,
    borderWidth: 1,
    borderColor: '#e5e7eb',
    padding: 12,
    minHeight: 150,
  },
});
