import * as ImagePicker from 'expo-image-picker';

/** Opens the photo library for a square crop. Resolves to the picked image's uri, or undefined if cancelled. */
export async function pickSquarePhoto(): Promise<string | undefined> {
  const res = await ImagePicker.launchImageLibraryAsync({ mediaTypes: ['images'], allowsEditing: true, aspect: [1, 1], quality: 0.7 });
  return res.canceled ? undefined : res.assets?.[0]?.uri;
}
