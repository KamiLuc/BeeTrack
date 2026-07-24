import 'package:flutter/material.dart';
import 'package:flutter_map/flutter_map.dart';
import 'package:latlong2/latlong.dart';

import '../../../core/widgets/profile_icon_button.dart';
import '../../../l10n/app_localizations.dart';
import '../data/apiary_model.dart';

class ApiariesMapScreen extends StatelessWidget {
  final List<Apiary> apiaries;

  /// Overrides the default "Apiaries map" AppBar title — useful when showing
  /// a single apiary (e.g. from a marketplace listing) where the plural
  /// title would read oddly.
  final String? title;

  const ApiariesMapScreen({super.key, required this.apiaries, this.title});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final located = apiaries
        .where((a) => a.lat != null && a.lng != null)
        .toList();

    final center = _computeCenter(located);

    final circles = <CircleMarker>[];
    for (final apiary in located) {
      final point = LatLng(apiary.lat!, apiary.lng!);
      circles.addAll([
        CircleMarker(
          point: point,
          radius: 5000,
          useRadiusInMeter: true,
          color: Colors.red.withAlpha(60),
          borderColor: Colors.red,
          borderStrokeWidth: 2,
        ),
        CircleMarker(
          point: point,
          radius: 3000,
          useRadiusInMeter: true,
          color: Colors.orange.withAlpha(60),
          borderColor: Colors.orange,
          borderStrokeWidth: 2,
        ),
        CircleMarker(
          point: point,
          radius: 1500,
          useRadiusInMeter: true,
          color: Colors.green.withAlpha(60),
          borderColor: Colors.green,
          borderStrokeWidth: 2,
        ),
      ]);
    }

    final markers = located
        .map((a) => Marker(
              point: LatLng(a.lat!, a.lng!),
              child: Tooltip(
                message: a.name,
                child: const Icon(Icons.location_pin,
                    color: Colors.amber, size: 36),
              ),
            ))
        .toList();

    return Scaffold(
      appBar: AppBar(
        title: Text(title ?? l10n.apiaryMapTitle),
        actions: const [ProfileIconButton()],
      ),
      body: FlutterMap(
        options: MapOptions(
          initialCenter: center,
          initialZoom: 10,
        ),
        children: [
          TileLayer(
            urlTemplate: 'https://tile.openstreetmap.org/{z}/{x}/{y}.png',
            userAgentPackageName: 'com.beetrack.app',
          ),
          CircleLayer(circles: circles),
          MarkerLayer(markers: markers),
        ],
      ),
    );
  }

  static LatLng _computeCenter(List<Apiary> located) {
    if (located.isEmpty) return const LatLng(52.2297, 21.0122);
    final lat = located.map((a) => a.lat!).reduce((a, b) => a + b) / located.length;
    final lng = located.map((a) => a.lng!).reduce((a, b) => a + b) / located.length;
    return LatLng(lat, lng);
  }
}
