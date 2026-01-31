import React from 'react';
import { View, Text, StyleSheet } from 'react-native';

interface StatsCardProps {
  countriesVisited: number;
  totalCountries: number;
  scrapbookEntries?: number;
}

export function StatsCard({ countriesVisited, totalCountries, scrapbookEntries }: StatsCardProps) {
  const percentage = totalCountries > 0
    ? Math.round((countriesVisited / totalCountries) * 100)
    : 0;

  return (
    <View style={styles.container}>
      <Text style={styles.title}>Your Journey</Text>

      <View style={styles.statsRow}>
        <View style={styles.statItem}>
          <Text style={styles.statValue}>{countriesVisited}</Text>
          <Text style={styles.statLabel}>Countries{'\n'}Visited</Text>
        </View>

        <View style={styles.divider} />

        <View style={styles.statItem}>
          <Text style={styles.statValue}>{percentage}%</Text>
          <Text style={styles.statLabel}>Of the{'\n'}World</Text>
        </View>

        {scrapbookEntries !== undefined && (
          <>
            <View style={styles.divider} />
            <View style={styles.statItem}>
              <Text style={styles.statValue}>{scrapbookEntries}</Text>
              <Text style={styles.statLabel}>Scrapbook{'\n'}Entries</Text>
            </View>
          </>
        )}
      </View>

      <View style={styles.progressContainer}>
        <View style={styles.progressBackground}>
          <View style={[styles.progressFill, { width: `${percentage}%` }]} />
        </View>
        <Text style={styles.progressText}>
          {countriesVisited} of {totalCountries} countries explored
        </Text>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    backgroundColor: '#4f46e5',
    borderRadius: 16,
    padding: 20,
    marginBottom: 16,
  },
  title: {
    fontSize: 18,
    fontWeight: '600',
    color: '#fff',
    marginBottom: 16,
  },
  statsRow: {
    flexDirection: 'row',
    justifyContent: 'space-around',
    alignItems: 'center',
    marginBottom: 20,
  },
  statItem: {
    alignItems: 'center',
    flex: 1,
  },
  statValue: {
    fontSize: 32,
    fontWeight: 'bold',
    color: '#fff',
  },
  statLabel: {
    fontSize: 12,
    color: 'rgba(255, 255, 255, 0.8)',
    textAlign: 'center',
    marginTop: 4,
  },
  divider: {
    width: 1,
    height: 40,
    backgroundColor: 'rgba(255, 255, 255, 0.3)',
  },
  progressContainer: {
    marginTop: 8,
  },
  progressBackground: {
    height: 8,
    backgroundColor: 'rgba(255, 255, 255, 0.3)',
    borderRadius: 4,
    overflow: 'hidden',
  },
  progressFill: {
    height: '100%',
    backgroundColor: '#22c55e',
    borderRadius: 4,
  },
  progressText: {
    fontSize: 12,
    color: 'rgba(255, 255, 255, 0.8)',
    textAlign: 'center',
    marginTop: 8,
  },
});

export default StatsCard;
