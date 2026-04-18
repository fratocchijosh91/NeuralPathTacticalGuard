import "../global.css";
import { Stack } from "expo-router";
import { StatusBar } from "expo-status-bar";
import { SafeAreaProvider } from "react-native-safe-area-context";

export default function RootLayout() {
  return (
    <SafeAreaProvider>
      <StatusBar style="light" />
      <Stack
        screenOptions={{
          headerStyle: { backgroundColor: "#030806" },
          headerTintColor: "#39ff14",
          headerShadowVisible: false,
          headerTitleStyle: { fontWeight: "700" },
          contentStyle: { backgroundColor: "#030806" },
        }}
      >
        <Stack.Screen name="index" options={{ title: "NeuralPath" }} />
        <Stack.Screen name="diagnostics" options={{ title: "Diagnostica" }} />
      </Stack>
    </SafeAreaProvider>
  );
}
