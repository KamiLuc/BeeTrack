import 'package:flutter/material.dart';
import '../../../core/widgets/profile_icon_button.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:geolocator/geolocator.dart';
import 'package:latlong2/latlong.dart';

import '../../../core/theme/app_layout.dart';
import '../../../l10n/app_localizations.dart';
import '../cubit/apiaries_cubit.dart';
import '../data/apiary_repository.dart';
import 'apiary_form_widgets.dart';
import 'map_picker_screen.dart';

class CreateApiaryScreen extends StatelessWidget {
  const CreateApiaryScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (_) => ApiariesCubit(repo: ApiaryRepository(api: context.read())),
      child: const _CreateApiaryView(),
    );
  }
}

class _CreateApiaryView extends StatefulWidget {
  const _CreateApiaryView();

  @override
  State<_CreateApiaryView> createState() => _CreateApiaryViewState();
}

class _CreateApiaryViewState extends State<_CreateApiaryView> {
  final _formKey = GlobalKey<FormState>();
  final _nameController = TextEditingController(text: 'Pasieka');
  final _latController = TextEditingController();
  final _lngController = TextEditingController();

  int _gridRows = 3;
  int _gridCols = 3;
  bool _locating = false;

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

  Future<void> _useGps(AppLocalizations l10n) async {
    setState(() => _locating = true);
    try {
      final serviceEnabled = await Geolocator.isLocationServiceEnabled();
      if (!serviceEnabled) {
        _showError(l10n.apiaryGpsUnavailable);
        return;
      }
      var permission = await Geolocator.checkPermission();
      if (permission == LocationPermission.denied) {
        permission = await Geolocator.requestPermission();
      }
      if (permission == LocationPermission.denied ||
          permission == LocationPermission.deniedForever) {
        _showError(l10n.apiaryGpsUnavailable);
        return;
      }
      final pos = await Geolocator.getCurrentPosition();
      setState(() => _setLocation(LatLng(pos.latitude, pos.longitude)));
    } catch (_) {
      _showError(l10n.apiaryGpsUnavailable);
    } finally {
      setState(() => _locating = false);
    }
  }

  Future<void> _pickOnMap() async {
    final result = await Navigator.of(context).push<LatLng>(
      MaterialPageRoute(builder: (_) => MapPickerScreen(initial: _location)),
    );
    if (result != null) setState(() => _setLocation(result));
  }

  void _showError(String message) {
    if (!mounted) return;
    ScaffoldMessenger.of(
      context,
    ).showSnackBar(SnackBar(content: Text(message)));
  }

  Future<void> _submit(AppLocalizations l10n) async {
    if (!(_formKey.currentState?.validate() ?? false)) return;
    await context.read<ApiariesCubit>().create(
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
        appBar: AppBar(title: Text(l10n.apiaryAdd), actions: const [ProfileIconButton()]),
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
                    autovalidateMode: AutovalidateMode.onUserInteraction,
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        TextFormField(
                          controller: _nameController,
                          decoration: InputDecoration(
                            labelText: l10n.apiaryName,
                          ),
                          textInputAction: TextInputAction.done,
                          validator: (v) => (v == null || v.trim().isEmpty)
                              ? l10n.apiaryNameRequired
                              : null,
                        ),
                        const SizedBox(height: 24),
                        ApiaryGridSection(
                          rows: _gridRows,
                          cols: _gridCols,
                          onRowsChanged: (v) => setState(() => _gridRows = v),
                          onColsChanged: (v) => setState(() => _gridCols = v),
                          l10n: l10n,
                        ),
                        const SizedBox(height: 24),
                        ApiaryLocationSection(
                          latController: _latController,
                          lngController: _lngController,
                          locating: _locating,
                          onGps: () => _useGps(l10n),
                          onMap: _pickOnMap,
                          l10n: l10n,
                        ),
                        const SizedBox(height: 16),
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
