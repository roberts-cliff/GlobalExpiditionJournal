import { useState } from 'react';
import { Tabs } from 'expo-router';
import { useAuth } from '../../context/AuthContext';
import { apiUrl } from '../../config/api';
import { ActivityIndicator, View, Text, StyleSheet, TouchableOpacity } from 'react-native';

export default function TabLayout() {
  const { isLoading, isAuthenticated, checkSession } = useAuth();
  const [demoLoading, setDemoLoading] = useState(false);

  const handleDemoLogin = async () => {
    setDemoLoading(true);
    try {
      const response = await fetch(apiUrl('/api/v1/demo/login'), {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: 'Demo Explorer', role: 'learner' }),
      });
      if (response.ok) {
        await checkSession();
      }
    } catch (error) {
      console.error('Demo login failed:', error);
    } finally {
      setDemoLoading(false);
    }
  };

  if (isLoading) {
    return (
      <View style={styles.centered}>
        <ActivityIndicator size="large" color="#4A90A4" />
        <Text style={styles.loadingText}>Loading...</Text>
      </View>
    );
  }

  if (!isAuthenticated) {
    return (
      <View style={styles.centered}>
        <Text style={styles.title}>Globe Expedition Journal</Text>
        <Text style={styles.subtitle}>Please launch from Canvas LMS</Text>
        <Text style={styles.orText}>— or —</Text>
        <TouchableOpacity
          style={styles.demoButton}
          onPress={handleDemoLogin}
          disabled={demoLoading}
        >
          {demoLoading ? (
            <ActivityIndicator color="#fff" />
          ) : (
            <Text style={styles.demoButtonText}>Try Demo Mode</Text>
          )}
        </TouchableOpacity>
      </View>
    );
  }

  return (
    <Tabs
      screenOptions={{
        tabBarActiveTintColor: '#4A90A4',
        tabBarInactiveTintColor: '#888',
        headerStyle: {
          backgroundColor: '#4A90A4',
        },
        headerTintColor: '#fff',
        headerTitleStyle: {
          fontWeight: 'bold',
        },
      }}
    >
      <Tabs.Screen
        name="passport"
        options={{
          title: 'Passport',
          tabBarLabel: 'Passport',
        }}
      />
      <Tabs.Screen
        name="scrapbook"
        options={{
          title: 'Scrapbook',
          tabBarLabel: 'Scrapbook',
        }}
      />
      <Tabs.Screen
        name="library"
        options={{
          title: 'Library',
          tabBarLabel: 'Library',
        }}
      />
      <Tabs.Screen
        name="checklist"
        options={{
          title: 'Checklist',
          tabBarLabel: 'Checklist',
        }}
      />
      <Tabs.Screen
        name="awards"
        options={{
          title: 'Awards',
          tabBarLabel: 'Awards',
        }}
      />
      <Tabs.Screen
        name="wishlist"
        options={{
          title: 'Wishlist',
          tabBarLabel: 'Wishlist',
        }}
      />
    </Tabs>
  );
}

const styles = StyleSheet.create({
  centered: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#f5f5f5',
  },
  loadingText: {
    marginTop: 16,
    fontSize: 16,
    color: '#666',
  },
  title: {
    fontSize: 28,
    fontWeight: 'bold',
    color: '#4A90A4',
    marginBottom: 8,
  },
  subtitle: {
    fontSize: 16,
    color: '#666',
  },
  orText: {
    marginTop: 24,
    marginBottom: 16,
    fontSize: 14,
    color: '#999',
  },
  demoButton: {
    backgroundColor: '#4A90A4',
    paddingHorizontal: 32,
    paddingVertical: 14,
    borderRadius: 8,
    minWidth: 160,
    alignItems: 'center',
  },
  demoButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
});
