import { Navigate, Route, Routes } from "react-router-dom";
import { AuthProvider } from "./auth/AuthContext";
import { RequireAuth } from "./auth/RequireAuth";
import { Layout } from "./components/Layout";
import { CertificationDetailPage } from "./pages/CertificationDetailPage";
import { CertificationQueuePage } from "./pages/CertificationQueuePage";
import { ListingDetailPage } from "./pages/ListingDetailPage";
import { ListingsQueuePage } from "./pages/ListingsQueuePage";
import { LoginPage } from "./pages/LoginPage";

export function App() {
  return (
    <AuthProvider>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route
          element={
            <RequireAuth>
              <Layout />
            </RequireAuth>
          }
        >
          <Route path="/listings" element={<ListingsQueuePage />} />
          <Route path="/listings/:id" element={<ListingDetailPage />} />
          <Route path="/certifications" element={<CertificationQueuePage />} />
          <Route path="/certifications/:id" element={<CertificationDetailPage />} />
        </Route>
        <Route path="*" element={<Navigate to="/listings" replace />} />
      </Routes>
    </AuthProvider>
  );
}
