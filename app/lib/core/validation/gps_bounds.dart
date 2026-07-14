/// Clamps a latitude to the valid range [-90, 90]. Mirrors the backend's
/// `validGPS` check in `internal/service/apiary.go`.
double clampLatitude(double lat) => lat.clamp(-90.0, 90.0);

/// Normalizes a longitude into the valid range [-180, 180], wrapping values
/// from panning a map across multiple world copies rather than clipping them.
double clampLongitude(double lng) {
  var normalized = lng % 360;
  if (normalized > 180) normalized -= 360;
  if (normalized < -180) normalized += 360;
  return normalized;
}
