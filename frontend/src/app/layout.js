import "@/styles/globals.css";
import { AuthProvider } from "@/context/authcontext";
import ToastContainer from "@/components/ui/ToastContainer";
import { WebSocketProvider } from "@/context/websocketContext";
import { UserStatusProvider } from "@/context/userStatusContext";
import { ChatProvider } from "@/context/chatContext";

export const metadata = {
  title: "Vibes",
  description: "Share your vibes with the world",
};

export default function RootLayout({ children }) {
  return (
    <html lang="en">
      <head>
        <link
          rel="stylesheet"
          href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.5.1/css/all.min.css"
        />
      </head>
      <body>
        <AuthProvider>
          <WebSocketProvider>
            <UserStatusProvider>
              <ChatProvider>{children}</ChatProvider>
            </UserStatusProvider>
            <ToastContainer />
          </WebSocketProvider>
        </AuthProvider>
      </body>
    </html>
  );
}
