import React from 'react';
import { View, Text, StyleSheet, TouchableOpacity, Image } from 'react-native';
import { ScrapbookEntry } from '../types/api';

interface ScrapbookEntryCardProps {
  entry: ScrapbookEntry;
  onPress: (entry: ScrapbookEntry) => void;
}

export function ScrapbookEntryCard({ entry, onPress }: ScrapbookEntryCardProps) {
  const formattedDate = entry.visitedAt
    ? new Date(entry.visitedAt).toLocaleDateString()
    : new Date(entry.createdAt).toLocaleDateString();

  return (
    <TouchableOpacity
      style={styles.container}
      onPress={() => onPress(entry)}
      activeOpacity={0.8}
    >
      {entry.mediaUrl ? (
        <Image source={{ uri: entry.mediaUrl }} style={styles.image} />
      ) : (
        <View style={styles.placeholderImage}>
          <Text style={styles.placeholderIcon}>📷</Text>
        </View>
      )}

      <View style={styles.content}>
        <Text style={styles.title} numberOfLines={1}>
          {entry.title}
        </Text>

        {entry.country && (
          <Text style={styles.country} numberOfLines={1}>
            {entry.country.name}
          </Text>
        )}

        <Text style={styles.date}>{formattedDate}</Text>

        {entry.notes && (
          <Text style={styles.notes} numberOfLines={2}>
            {entry.notes}
          </Text>
        )}

        {entry.tags && (
          <View style={styles.tagsContainer}>
            {entry.tags.split(',').slice(0, 3).map((tag, index) => (
              <View key={index} style={styles.tag}>
                <Text style={styles.tagText}>{tag.trim()}</Text>
              </View>
            ))}
            {entry.tags.split(',').length > 3 && (
              <Text style={styles.moreTagsText}>+{entry.tags.split(',').length - 3}</Text>
            )}
          </View>
        )}
      </View>
    </TouchableOpacity>
  );
}

const styles = StyleSheet.create({
  container: {
    backgroundColor: '#fff',
    borderRadius: 12,
    overflow: 'hidden',
    marginBottom: 12,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
  },
  image: {
    width: '100%',
    height: 160,
    backgroundColor: '#e5e7eb',
  },
  placeholderImage: {
    width: '100%',
    height: 160,
    backgroundColor: '#f3f4f6',
    justifyContent: 'center',
    alignItems: 'center',
  },
  placeholderIcon: {
    fontSize: 48,
    opacity: 0.5,
  },
  content: {
    padding: 12,
  },
  title: {
    fontSize: 16,
    fontWeight: '600',
    color: '#111827',
    marginBottom: 4,
  },
  country: {
    fontSize: 14,
    color: '#4f46e5',
    marginBottom: 4,
  },
  date: {
    fontSize: 12,
    color: '#9ca3af',
    marginBottom: 8,
  },
  notes: {
    fontSize: 14,
    color: '#6b7280',
    lineHeight: 20,
  },
  tagsContainer: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    marginTop: 8,
    gap: 6,
    alignItems: 'center',
  },
  tag: {
    backgroundColor: '#f3f4f6',
    paddingHorizontal: 8,
    paddingVertical: 3,
    borderRadius: 10,
  },
  tagText: {
    fontSize: 11,
    color: '#6b7280',
  },
  moreTagsText: {
    fontSize: 11,
    color: '#9ca3af',
    marginLeft: 4,
  },
});

export default ScrapbookEntryCard;
