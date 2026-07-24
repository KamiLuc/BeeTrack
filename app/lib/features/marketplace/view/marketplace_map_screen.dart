import 'package:flutter/material.dart';
import 'package:flutter_map/flutter_map.dart';
import 'package:latlong2/latlong.dart';

import '../../../core/widgets/profile_icon_button.dart';
import '../../../l10n/app_localizations.dart';
import '../data/listing_model.dart';
import '../data/listing_price.dart';
import 'listing_detail_screen.dart';

/// Shows the listings passed in (already filtered/searched by
/// [MarketplaceCubit] on the home screen) as pins on a map. Listings without
/// a real location (lat/lng both 0) are excluded rather than pinned at
/// "null island".
class MarketplaceMapScreen extends StatelessWidget {
  final List<Listing> listings;

  const MarketplaceMapScreen({super.key, required this.listings});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final located = listings
        .where((listing) => listing.lat != 0 || listing.lng != 0)
        .toList();

    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.marketplaceMapTitle),
        actions: const [ProfileIconButton()],
      ),
      body: located.isEmpty
          ? Center(child: Text(l10n.marketplaceMapEmpty))
          : FlutterMap(
              options: MapOptions(
                initialCenter: _computeCenter(located),
                initialZoom: 10,
              ),
              children: [
                TileLayer(
                  urlTemplate: 'https://tile.openstreetmap.org/{z}/{x}/{y}.png',
                  userAgentPackageName: 'com.beetrack.app',
                ),
                MarkerLayer(
                  markers: located
                      .map((listing) => _buildMarker(context, l10n, listing))
                      .toList(),
                ),
              ],
            ),
    );
  }

  Marker _buildMarker(
    BuildContext context,
    AppLocalizations l10n,
    Listing listing,
  ) {
    return Marker(
      point: LatLng(listing.lat, listing.lng),
      child: GestureDetector(
        onTap: () => Navigator.of(context).push(
          MaterialPageRoute(
            builder: (_) => ListingDetailScreen(listing: listing),
          ),
        ),
        child: Tooltip(
          message:
              '${listing.title} • ${listingPriceLabel(l10n, listing.price)}',
          child: const Icon(Icons.location_pin, color: Colors.amber, size: 36),
        ),
      ),
    );
  }

  static LatLng _computeCenter(List<Listing> located) {
    if (located.isEmpty) return const LatLng(52.2297, 21.0122);
    final lat = located.map((l) => l.lat).reduce((a, b) => a + b) / located.length;
    final lng = located.map((l) => l.lng).reduce((a, b) => a + b) / located.length;
    return LatLng(lat, lng);
  }
}
