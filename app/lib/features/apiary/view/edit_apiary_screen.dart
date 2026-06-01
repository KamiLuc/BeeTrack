import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:latlong2/latlong.dart';

import '../../../core/theme/app_layout.dart';
import '../../../l10n/app_localizations.dart';
import '../cubit/apiaries_cubit.dart';
import '../data/apiary_model.dart';
import '../data/apiary_repository.dart';
import 'create_apiary_screen.dart';
import 'map_picker_screen.dart';

class EditApiaryScreen extends StatelessWidget {
  final Apiary apiary;

  const EditApiaryScreen({super.key, required this.apiary});

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (_) => ApiariesCubit(
        repo: ApiaryRepository(api: context.read()),
      ),
      child: _EditApiaryView(apiary: apiary),
    );
  }
}

class _EditApiaryView extends StatefulWidget {
  final Apiary apiary;

  const _EditApiaryView({required this.apiary});

  @override
  State<_EditApiaryView> createState() => _EditApiaryViewState();
}

class _EditApiaryViewState extends State<_EditApiaryView> {
  final _formKey = GlobalKey<FormState>();
  late final TextEditingController _nameController;
  late final TextEditingController _latController;
  late final TextEditingController _lngController;

  late int _gridRows;
  late int _gridCols;

  @override
  void initState() {
    super.initState();
    _nameController = TextEditingController(text: widget.apiary.name);
    _latController = TextEditingController(
      text: widget.apiary.lat?.toStringAsFixed(6) ?? '',
    );
    _lngController = TextEditingController(
      text: widget.apiary.lng?.toStringAsFixed(6) ?? '',
    );
    _gridRows = widget.apiary.gridRows;
    _gridCols = widget.apiary.gridCols;
  }

  @override
  void dispose() {
    _nameController.dispose();
    _latController.dispose();
    _lngController.dispose();
    super.dispose();
  }

  LatLng? get _location {
    final lat = double.tryParse(_latController.text);
    final lng = double.tryParse(_lngController.text);
    if (lat != null && lng != null) return LatLng(lat, lng);
    return null;
  }

  void _setLocation(LatLng loc) {
    _latController.text = loc.latitude.toStringAsFixed(6);
    _lngController.text = loc.longitude.toStringAsFixed(6);
  }

  Future<void> _pickOnMap() async {
    final result = await Navigator.of(context).push<LatLng>(
      MaterialPageRoute(
        builder: (_) => MapPickerScreen(initial: _location),
      ),
    );
    if (result != null) setState(() => _setLocation(result));
  }

  void _showError(String message) {
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(message)),
    );
  }

  Future<void> _submit(AppLocalizations l10n) async {
    if (!(_formKey.currentState?.validate() ?? false)) return;
    await context.read<ApiariesCubit>().update(
      id: widget.apiary.id,
      name: _nameController.text.trim(),
      lat: _location?.latitude,
      lng: _location?.longitude,
      gridRows: _gridRows,
      gridCols: _gridCols,
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return BlocListener<ApiariesCubit, ApiariesState>(
      listener: (context, state) {
        if (state is ApiariesLoaded) Navigator.of(context).pop();
        if (state is ApiariesError) _showError(l10n.generalError);
      },
      child: Scaffold(
        appBar: AppBar(title: Text(l10n.apiaryEdit)),
        body: BlocBuilder<ApiariesCubit, ApiariesState>(
          builder: (context, state) {
            final loading = state is ApiariesLoading;
            return Center(
              child: SingleChildScrollView(
                padding: const EdgeInsets.all(24),
                child: ConstrainedBox(
                  constraints: AppLayout.formConstraints(context),
                  child: Form(
                    key: _formKey,
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        TextFormField(
                          controller: _nameController,
                          decoration: InputDecoration(labelText: l10n.apiaryName),
                          textInputAction: TextInputAction.done,
                          validator: (v) => (v == null || v.trim().isEmpty)
                              ? l10n.apiaryName
                              : null,
                        ),
                        const SizedBox(height: 24),
                        _GridSection(
                          rows: _gridRows,
                          cols: _gridCols,
                          onRowsChanged: (v) => setState(() => _gridRows = v),
                          onColsChanged: (v) => setState(() => _gridCols = v),
                          l10n: l10n,
                        ),
                        const SizedBox(height: 24),
                        _LocationSection(
                          latController: _latController,
                          lngController: _lngController,
                          onMap: _pickOnMap,
                          l10n: l10n,
                        ),
                        const SizedBox(height: 32),
                        Center(
                          child: SizedBox(
                            width: 200,
                            child: ElevatedButton(
                              onPressed: loading ? null : () => _submit(l10n),
                              child: loading
                                  ? const SizedBox(
                                      height: 20,
                                      width: 20,
                                      child: CircularProgressIndicator(
                                        strokeWidth: 2,
                                        color: Colors.white,
                                      ),
                                    )
                                  : Text(l10n.generalSave),
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ),
            );
          },
        ),
      ),
    );
  }
}

class _GridSection extends StatelessWidget {
  final int rows;
  final int cols;
  final ValueChanged<int> onRowsChanged;
  final ValueChanged<int> onColsChanged;
  final AppLocalizations l10n;

  const _GridSection({
    required this.rows,
    required this.cols,
    required this.onRowsChanged,
    required this.onColsChanged,
    required this.l10n,
  });

  @override
  Widget build(BuildContext context) {
    final items = List.generate(25, (i) => i + 1)
        .map((v) => DropdownMenuItem(value: v, child: Text('$v')))
        .toList();

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Row(
          children: [
            Expanded(
              child: DropdownButtonFormField<int>(
                value: rows,
                menuMaxHeight: 48 * 10,
                decoration: InputDecoration(labelText: l10n.apiaryGridRows),
                items: items,
                onChanged: (v) { if (v != null) onRowsChanged(v); },
              ),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: DropdownButtonFormField<int>(
                value: cols,
                menuMaxHeight: 48 * 10,
                decoration: InputDecoration(labelText: l10n.apiaryGridCols),
                items: items,
                onChanged: (v) { if (v != null) onColsChanged(v); },
              ),
            ),
          ],
        ),
        const SizedBox(height: 16),
        _GridPreview(rows: rows, cols: cols),
      ],
    );
  }
}

class _GridPreview extends StatelessWidget {
  final int rows;
  final int cols;

  const _GridPreview({required this.rows, required this.cols});

  @override
  Widget build(BuildContext context) {
    const maxWidth = 240.0;
    final cellSize = (maxWidth / cols).clamp(4.0, 24.0);

    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: List.generate(rows, (r) {
          return Row(
            mainAxisSize: MainAxisSize.min,
            children: List.generate(cols, (c) {
              return Container(
                width: cellSize,
                height: cellSize,
                margin: const EdgeInsets.all(1),
                decoration: BoxDecoration(
                  color: Theme.of(context).colorScheme.primaryContainer,
                  borderRadius: BorderRadius.circular(2),
                ),
              );
            }),
          );
        }),
      ),
    );
  }
}

class _LocationSection extends StatelessWidget {
  final TextEditingController latController;
  final TextEditingController lngController;
  final VoidCallback onMap;
  final AppLocalizations l10n;

  const _LocationSection({
    required this.latController,
    required this.lngController,
    required this.onMap,
    required this.l10n,
  });

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        TextFormField(
          controller: latController,
          enabled: false,
          decoration: InputDecoration(labelText: l10n.apiaryLatitude),
        ),
        const SizedBox(height: 12),
        TextFormField(
          controller: lngController,
          enabled: false,
          decoration: InputDecoration(labelText: l10n.apiaryLongitude),
        ),
        const SizedBox(height: 12),
        OutlinedButton.icon(
          onPressed: onMap,
          icon: const Icon(Icons.map, size: 18),
          label: const Text('Mapa'),
        ),
      ],
    );
  }
}
