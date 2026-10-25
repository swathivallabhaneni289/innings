import { useEffect, useState, type ReactNode } from 'react';
import { Dimensions, Keyboard, KeyboardAvoidingView, Platform, View, type ViewStyle } from 'react-native';

/**
 * Keeps a screen's content above the keyboard. iOS uses KeyboardAvoidingView. Android draws edge to edge, so the
 * window no longer resizes for the keyboard and KeyboardAvoidingView only lifts part of it: there the content is
 * padded from the keyboard's top edge to the bottom of the screen (its reported height leaves out the navigation
 * bar under it), and a ScrollView inside scrolls the focused field into view.
 */
export function KeyboardAware({ children, style }: { children: ReactNode; style?: ViewStyle }) {
  const height = useAndroidKeyboardHeight();
  if (Platform.OS === 'ios') {
    return (
      <KeyboardAvoidingView behavior="padding" style={[{ flex: 1 }, style]}>
        {children}
      </KeyboardAvoidingView>
    );
  }
  return <View style={[{ flex: 1, paddingBottom: height }, style]}>{children}</View>;
}

function useAndroidKeyboardHeight() {
  const [height, setHeight] = useState(0);
  useEffect(() => {
    if (Platform.OS !== 'android') return;
    const show = Keyboard.addListener('keyboardDidShow', (e) =>
      setHeight(Math.max(e.endCoordinates.height, Dimensions.get('screen').height - e.endCoordinates.screenY)),
    );
    const hide = Keyboard.addListener('keyboardDidHide', () => setHeight(0));
    return () => {
      show.remove();
      hide.remove();
    };
  }, []);
  return height;
}
