// requestGeolocation resolves to a partial StartQuizSubmission payload carrying
// either GPS coordinates or a gps_denied flag. It never rejects: on missing API,
// permission denial, error, or timeout it resolves with { gps_denied: true }.
// The backend requires exactly one of the two shapes when require_gps is on.
export interface GeolocationResult {
  gps_lat?: number
  gps_lng?: number
  gps_accuracy?: number
  gps_denied?: boolean
}

export function requestGeolocation(): Promise<GeolocationResult> {
  return new Promise((resolve) => {
    if (typeof navigator === "undefined" || !("geolocation" in navigator)) {
      resolve({ gps_denied: true })
      return
    }
    navigator.geolocation.getCurrentPosition(
      (pos) =>
        resolve({
          gps_lat: pos.coords.latitude,
          gps_lng: pos.coords.longitude,
          gps_accuracy: pos.coords.accuracy,
        }),
      () => resolve({ gps_denied: true }),
      { enableHighAccuracy: true, timeout: 10000, maximumAge: 0 },
    )
  })
}
