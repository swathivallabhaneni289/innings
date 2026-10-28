import { useEffect, useState, type ReactNode } from 'react';
import { Dimensions, Keyboard, KeyboardAvoidingView, Platform, View, type ViewStyle } from 'react-native';

/**
 * Keeps a screen's content above the keyboard. Wrap the whole screen, header included: on iOS this is a
 * KeyboardAvoidingView, which measures itself against its parent, so one placed below a header lifts too little.
 * Android draws edge to edge, so the window no longer resizes for the keyboard and KeyboardAvoidingView only lifts
 * part of it: there the content is padded from the keyboard's top edge to the bottom of the screen (its reported
 * height leaves out the navigation bar under it), and a ScrollView inside scrolls the focused field into view.
 *
 * iOS doesn't scroll to the focused field by itself. A screen that is one ScrollView of fields passes `scrollInsets`
 * and sets `automaticallyAdjustKeyboardInsets` on the ScrollView, which then insets itself and scrolls the field
 * above the keyboard; KeyboardAvoidingView is left out there, or the two would pad twice.
 */
export function KeyboardAware({
  children,
  style,
  scrollInsets = false,
}: {
  children: ReactNode;
  style?: ViewStyle;
  scrollInsets?: boolean;
}) {
  const height = useAndroidKeyboardHeight();
  if (Platform.OS === 'ios') {
    if (scrollInsets) return <View style={[{ flex: 1 }, style]}>{children}</View>;
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
