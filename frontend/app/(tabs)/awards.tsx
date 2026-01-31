import { View, Text, StyleSheet } from 'react-native';

export default function AwardsScreen() {
  return (
    <View style={styles.container}>
      <Text style={styles.title}>My Awards</Text>
      <View style={styles.placeholder}>
        <Text style={styles.placeholderText}>
          Earn badges and achievements as you explore the world
        </Text>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    padding: 16,
    backgroundColor: '#f5f5f5',
  },
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#333',
    marginBottom: 24,
  },
  placeholder: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 24,
  },
  placeholderText: {
    fontSize: 16,
    color: '#999',
    textAlign: 'center',
  },
});
