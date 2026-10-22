import { useLocalSearchParams } from 'expo-router';

import { ProfileView } from '@/features/profile/ProfileView';

export default function FanScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  return <ProfileView id={id} />;
}
