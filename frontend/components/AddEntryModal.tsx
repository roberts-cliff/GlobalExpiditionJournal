import React, { useState } from 'react';
import {
  View,
  Text,
  StyleSheet,
  Modal,
  TextInput,
  TouchableOpacity,
  ScrollView,
  KeyboardAvoidingView,
  Platform,
  Image,
  Alert,
  ActivityIndicator,
} from 'react-native';
import * as ImagePicker from 'expo-image-picker';
import { Country, CreateScrapbookEntryRequest } from '../types/api';
import { API_BASE_URL } from '../config/api';

interface AddEntryModalProps {
  visible: boolean;
  onClose: () => void;
  onSave: (entry: CreateScrapbookEntryRequest) => Promise<void>;
  countries: Country[];
  loading?: boolean;
}

export function AddEntryModal({
  visible,
  onClose,
  onSave,
  countries,
  loading = false,
}: AddEntryModalProps) {
  const [selectedCountry, setSelectedCountry] = useState<Country | null>(null);
  const [title, setTitle] = useState('');
  const [notes, setNotes] = useState('');
  const [tags, setTags] = useState('');
  const [showCountryPicker, setShowCountryPicker] = useState(false);
  const [selectedImage, setSelectedImage] = useState<string | null>(null);
  const [uploadedMediaUrl, setUploadedMediaUrl] = useState<string | null>(null);
  const [uploading, setUploading] = useState(false);

  const resetForm = () => {
    setSelectedCountry(null);
    setTitle('');
    setNotes('');
    setTags('');
    setSelectedImage(null);
    setUploadedMediaUrl(null);
  };

  const handleClose = () => {
    resetForm();
    onClose();
  };

  const pickImage = async () => {
    // Ask for permission
    const permissionResult = await ImagePicker.requestMediaLibraryPermissionsAsync();

    if (!permissionResult.granted) {
      Alert.alert(
        'Permission Required',
        'Please allow access to your photo library to add photos to your memories.'
      );
      return;
    }

    // Launch image picker
    const result = await ImagePicker.launchImageLibraryAsync({
      mediaTypes: ['images'],
      allowsEditing: true,
      aspect: [4, 3],
      quality: 0.8,
    });

    if (!result.canceled && result.assets[0]) {
      const asset = result.assets[0];
      setSelectedImage(asset.uri);

      // Upload the image
      await uploadImage(asset.uri, asset.mimeType || 'image/jpeg');
    }
  };

  const takePhoto = async () => {
    // Ask for camera permission
    const permissionResult = await ImagePicker.requestCameraPermissionsAsync();

    if (!permissionResult.granted) {
      Alert.alert(
        'Permission Required',
        'Please allow access to your camera to take photos for your memories.'
      );
      return;
    }

    // Launch camera
    const result = await ImagePicker.launchCameraAsync({
      allowsEditing: true,
      aspect: [4, 3],
      quality: 0.8,
    });

    if (!result.canceled && result.assets[0]) {
      const asset = result.assets[0];
      setSelectedImage(asset.uri);

      // Upload the image
      await uploadImage(asset.uri, asset.mimeType || 'image/jpeg');
    }
  };

  const uploadImage = async (uri: string, mimeType: string) => {
    setUploading(true);

    try {
      // Create form data for upload
      const formData = new FormData();

      // Extract filename from URI
      const filename = uri.split('/').pop() || 'photo.jpg';

      // Append file to form data
      formData.append('file', {
        uri: uri,
        name: filename,
        type: mimeType,
      } as any);

      const response = await fetch(`${API_BASE_URL}/api/v1/upload`, {
        method: 'POST',
        body: formData,
        headers: {
          'Content-Type': 'multipart/form-data',
        },
        credentials: 'include',
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || 'Upload failed');
      }

      const data = await response.json();
      setUploadedMediaUrl(data.url);
    } catch (error) {
      console.error('Upload error:', error);
      Alert.alert(
        'Upload Failed',
        error instanceof Error ? error.message : 'Failed to upload photo. Please try again.'
      );
      setSelectedImage(null);
    } finally {
      setUploading(false);
    }
  };

  const removePhoto = () => {
    setSelectedImage(null);
    setUploadedMediaUrl(null);
  };

  const showPhotoOptions = () => {
    Alert.alert(
      'Add Photo',
      'Choose how you want to add a photo',
      [
        { text: 'Take Photo', onPress: takePhoto },
        { text: 'Choose from Library', onPress: pickImage },
        { text: 'Cancel', style: 'cancel' },
      ]
    );
  };

  const handleSave = async () => {
    if (!selectedCountry || !title.trim()) return;

    // Clean up tags: trim whitespace, remove empty tags
    const cleanedTags = tags
      .split(',')
      .map(t => t.trim())
      .filter(t => t.length > 0)
      .join(',');

    await onSave({
      countryId: selectedCountry.id,
      title: title.trim(),
      notes: notes.trim() || undefined,
      tags: cleanedTags || undefined,
      mediaUrl: uploadedMediaUrl || undefined,
      mediaType: uploadedMediaUrl ? 'image/jpeg' : undefined,
    });

    resetForm();
  };

  const isValid = selectedCountry && title.trim().length > 0;
  const isBusy = loading || uploading;

  return (
    <Modal
      visible={visible}
      animationType="slide"
      presentationStyle="pageSheet"
      onRequestClose={handleClose}
    >
      <KeyboardAvoidingView
        behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
        style={styles.container}
      >
        <View style={styles.header}>
          <TouchableOpacity onPress={handleClose} style={styles.headerButton}>
            <Text style={styles.cancelText}>Cancel</Text>
          </TouchableOpacity>

          <Text style={styles.headerTitle}>New Memory</Text>

          <TouchableOpacity
            onPress={handleSave}
            style={styles.headerButton}
            disabled={!isValid || isBusy}
          >
            <Text
              style={[
                styles.saveText,
                (!isValid || isBusy) && styles.saveTextDisabled,
              ]}
            >
              {loading ? 'Saving...' : 'Save'}
            </Text>
          </TouchableOpacity>
        </View>

        <ScrollView style={styles.content} keyboardShouldPersistTaps="handled">
          {/* Country Picker */}
          <Text style={styles.label}>Country *</Text>
          <TouchableOpacity
            style={styles.pickerButton}
            onPress={() => setShowCountryPicker(!showCountryPicker)}
          >
            <Text
              style={[
                styles.pickerButtonText,
                !selectedCountry && styles.pickerPlaceholder,
              ]}
            >
              {selectedCountry?.name || 'Select a country'}
            </Text>
            <Text style={styles.pickerArrow}>
              {showCountryPicker ? '▲' : '▼'}
            </Text>
          </TouchableOpacity>

          {showCountryPicker && (
            <View style={styles.countryList}>
              <ScrollView style={styles.countryScroll} nestedScrollEnabled>
                {countries.map((country) => (
                  <TouchableOpacity
                    key={country.id}
                    style={[
                      styles.countryItem,
                      selectedCountry?.id === country.id &&
                        styles.countryItemSelected,
                    ]}
                    onPress={() => {
                      setSelectedCountry(country);
                      setShowCountryPicker(false);
                    }}
                  >
                    <Text style={styles.countryCode}>{country.isoCode}</Text>
                    <Text style={styles.countryName}>{country.name}</Text>
                  </TouchableOpacity>
                ))}
              </ScrollView>
            </View>
          )}

          {/* Title */}
          <Text style={styles.label}>Title *</Text>
          <TextInput
            style={styles.input}
            placeholder="Give this memory a title"
            placeholderTextColor="#9ca3af"
            value={title}
            onChangeText={setTitle}
            maxLength={100}
          />

          {/* Notes */}
          <Text style={styles.label}>Notes</Text>
          <TextInput
            style={[styles.input, styles.textArea]}
            placeholder="Write about this memory..."
            placeholderTextColor="#9ca3af"
            value={notes}
            onChangeText={setNotes}
            multiline
            numberOfLines={4}
            textAlignVertical="top"
          />

          {/* Tags */}
          <Text style={styles.label}>Tags</Text>
          <TextInput
            style={styles.input}
            placeholder="museum, food, nature (comma separated)"
            placeholderTextColor="#9ca3af"
            value={tags}
            onChangeText={setTags}
            autoCapitalize="none"
          />
          {tags.length > 0 && (
            <View style={styles.tagPreview}>
              {tags.split(',').filter(t => t.trim()).map((tag, index) => (
                <View key={index} style={styles.tagChip}>
                  <Text style={styles.tagChipText}>{tag.trim()}</Text>
                </View>
              ))}
            </View>
          )}

          {/* Photo picker */}
          <Text style={styles.label}>Photo</Text>
          {selectedImage ? (
            <View style={styles.imagePreviewContainer}>
              <Image source={{ uri: selectedImage }} style={styles.imagePreview} />
              {uploading && (
                <View style={styles.uploadingOverlay}>
                  <ActivityIndicator size="large" color="#fff" />
                  <Text style={styles.uploadingText}>Uploading...</Text>
                </View>
              )}
              {!uploading && (
                <TouchableOpacity style={styles.removePhotoButton} onPress={removePhoto}>
                  <Text style={styles.removePhotoText}>✕</Text>
                </TouchableOpacity>
              )}
              {uploadedMediaUrl && !uploading && (
                <View style={styles.uploadedBadge}>
                  <Text style={styles.uploadedBadgeText}>✓ Uploaded</Text>
                </View>
              )}
            </View>
          ) : (
            <TouchableOpacity style={styles.photoButton} onPress={showPhotoOptions}>
              <Text style={styles.photoButtonIcon}>📷</Text>
              <Text style={styles.photoButtonText}>Add a photo</Text>
            </TouchableOpacity>
          )}
        </ScrollView>
      </KeyboardAvoidingView>
    </Modal>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#fff',
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: 16,
    paddingVertical: 12,
    borderBottomWidth: 1,
    borderBottomColor: '#e5e7eb',
  },
  headerButton: {
    minWidth: 60,
  },
  headerTitle: {
    fontSize: 17,
    fontWeight: '600',
    color: '#111827',
  },
  cancelText: {
    fontSize: 16,
    color: '#6b7280',
  },
  saveText: {
    fontSize: 16,
    fontWeight: '600',
    color: '#4f46e5',
    textAlign: 'right',
  },
  saveTextDisabled: {
    color: '#9ca3af',
  },
  content: {
    flex: 1,
    padding: 16,
  },
  label: {
    fontSize: 14,
    fontWeight: '600',
    color: '#374151',
    marginBottom: 8,
    marginTop: 16,
  },
  input: {
    backgroundColor: '#f9fafb',
    borderRadius: 8,
    borderWidth: 1,
    borderColor: '#e5e7eb',
    paddingHorizontal: 12,
    paddingVertical: 12,
    fontSize: 16,
    color: '#111827',
  },
  textArea: {
    minHeight: 100,
  },
  pickerButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    backgroundColor: '#f9fafb',
    borderRadius: 8,
    borderWidth: 1,
    borderColor: '#e5e7eb',
    paddingHorizontal: 12,
    paddingVertical: 12,
  },
  pickerButtonText: {
    fontSize: 16,
    color: '#111827',
  },
  pickerPlaceholder: {
    color: '#9ca3af',
  },
  pickerArrow: {
    fontSize: 12,
    color: '#6b7280',
  },
  countryList: {
    marginTop: 8,
    backgroundColor: '#f9fafb',
    borderRadius: 8,
    borderWidth: 1,
    borderColor: '#e5e7eb',
    maxHeight: 200,
  },
  countryScroll: {
    maxHeight: 200,
  },
  countryItem: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 12,
    paddingVertical: 10,
    borderBottomWidth: 1,
    borderBottomColor: '#e5e7eb',
  },
  countryItemSelected: {
    backgroundColor: '#eef2ff',
  },
  countryCode: {
    fontSize: 14,
    fontWeight: '600',
    color: '#6b7280',
    width: 40,
  },
  countryName: {
    fontSize: 16,
    color: '#111827',
  },
  photoButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: '#f9fafb',
    borderRadius: 8,
    borderWidth: 2,
    borderColor: '#e5e7eb',
    borderStyle: 'dashed',
    paddingVertical: 24,
  },
  photoButtonIcon: {
    fontSize: 24,
    marginRight: 8,
  },
  photoButtonText: {
    fontSize: 16,
    color: '#6b7280',
  },
  imagePreviewContainer: {
    position: 'relative',
    borderRadius: 8,
    overflow: 'hidden',
  },
  imagePreview: {
    width: '100%',
    height: 200,
    borderRadius: 8,
  },
  uploadingOverlay: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  uploadingText: {
    color: '#fff',
    marginTop: 8,
    fontSize: 14,
  },
  removePhotoButton: {
    position: 'absolute',
    top: 8,
    right: 8,
    backgroundColor: 'rgba(0, 0, 0, 0.6)',
    width: 28,
    height: 28,
    borderRadius: 14,
    justifyContent: 'center',
    alignItems: 'center',
  },
  removePhotoText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
  uploadedBadge: {
    position: 'absolute',
    bottom: 8,
    left: 8,
    backgroundColor: 'rgba(34, 197, 94, 0.9)',
    paddingHorizontal: 8,
    paddingVertical: 4,
    borderRadius: 4,
  },
  uploadedBadgeText: {
    color: '#fff',
    fontSize: 12,
    fontWeight: '600',
  },
  tagPreview: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    marginTop: 8,
    gap: 6,
  },
  tagChip: {
    backgroundColor: '#eef2ff',
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 12,
  },
  tagChipText: {
    fontSize: 12,
    color: '#4f46e5',
  },
});

export default AddEntryModal;
