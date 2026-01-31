import React from 'react';
import { View, Text, StyleSheet, TouchableOpacity } from 'react-native';
import { Country } from '../types/api';

interface CountryListItemProps {
  country: Country;
  isVisited: boolean;
  visitDate?: string;
  onPress: (country: Country) => void;
}

export function CountryListItem({ country, isVisited, visitDate, onPress }: CountryListItemProps) {
  return (
    <TouchableOpacity
      style={[styles.container, isVisited && styles.visitedContainer]}
      onPress={() => onPress(country)}
      activeOpacity={0.7}
    >
      <View style={styles.leftSection}>
        <Text style={styles.countryCode}>{country.isoCode}</Text>
      </View>

      <View style={styles.middleSection}>
        <Text style={styles.countryName}>{country.name}</Text>
        <Text style={styles.region}>{country.region}</Text>
      </View>

      <View style={styles.rightSection}>
        {isVisited ? (
          <View style={styles.visitedBadge}>
            <Text style={styles.visitedIcon}>✓</Text>
            {visitDate && (
              <Text style={styles.visitDate}>
                {new Date(visitDate).toLocaleDateString()}
              </Text>
            )}
          </View>
        ) : (
          <Text style={styles.notVisited}>Not visited</Text>
        )}
      </View>
    </TouchableOpacity>
  );
}

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 16,
    marginBottom: 8,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.1,
    shadowRadius: 2,
    elevation: 2,
  },
  visitedContainer: {
    backgroundColor: '#f0fdf4',
    borderLeftWidth: 4,
    borderLeftColor: '#22c55e',
  },
  leftSection: {
    width: 48,
    height: 48,
    borderRadius: 24,
    backgroundColor: '#e5e7eb',
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 12,
  },
  countryCode: {
    fontSize: 14,
    fontWeight: '600',
    color: '#374151',
  },
  middleSection: {
    flex: 1,
  },
  countryName: {
    fontSize: 16,
    fontWeight: '600',
    color: '#111827',
    marginBottom: 2,
  },
  region: {
    fontSize: 14,
    color: '#6b7280',
  },
  rightSection: {
    alignItems: 'flex-end',
  },
  visitedBadge: {
    alignItems: 'center',
  },
  visitedIcon: {
    fontSize: 20,
    color: '#22c55e',
    fontWeight: 'bold',
  },
  visitDate: {
    fontSize: 11,
    color: '#22c55e',
    marginTop: 2,
  },
  notVisited: {
    fontSize: 12,
    color: '#9ca3af',
  },
});

export default CountryListItem;
