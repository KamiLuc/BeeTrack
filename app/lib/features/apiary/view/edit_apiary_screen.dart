import 'package:flutter/material.dart';
import '../../../core/widgets/profile_icon_button.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:geolocator/geolocator.dart';
import 'package:latlong2/latlong.dart';

import '../../../core/theme/app_layout.dart';
import '../../../core/validation/gps_bounds.dart';
import '../../../core/validation/size_tiers.dart';
import '../../../l10n/app_localizations.dart';
import '../../hive/data/hive_model.dart';
import '../../hive/data/hive_repository.dart';
import '../cubit/apiaries_cubit.dart';
import '../data/apiary_model.dart';
import '../data/apiary_repository.dart';
import 'apiary_form_widgets.dart';
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
  bool _locating = false;
  List<Hive>? _hives;

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
    WidgetsBinding.instance.addPostFrameCallback((_) => _loadHives());
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
    _latController.text = clampLatitude(loc.latitude).toStringAsFixed(6);
    _lngController.text = clampLongitude(loc.longitude).toStringAsFixed(6);
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

  Future<void> _loadHives() async {
    try {
      final hives = await HiveRepository(api: context.read())
          .listHives(widget.apiary.id);
      if (mounted) setState(() => _hives = hives);
    } catch (_) {}
  }

  bool get _gridTooSmall =>
      _gridRows * _gridCols < widget.apiary.hiveCount;

  int get _outOfBoundsCount {
    if (_hives == null) return 0;
    return _hives!
        .where((h) => h.gridRow >= _gridRows || h.gridCol >= _gridCols)
        .length;
  }

  Future<void> _submit(AppLocalizations l10n) async {
    if (_gridTooSmall) return;
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
        if (state is ApiariesError) {
          final msg = state.code == 'GRID_TOO_SMALL'
              ? l10n.apiaryGridTooSmall
              : l10n.generalError;
          _showError(msg);
        }
      },
      child: Scaffold(
        appBar: AppBar(title: Text(l10n.apiaryEdit), actions: const [ProfileIconButton()]),
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
                            counterText: SizeTier.small.counterText,
                          ),
                          textInputAction: TextInputAction.done,
                          maxLength: SizeTier.small.maxLength,
                          validator: (v) {
                            if (v == null || v.trim().isEmpty) {
                              return l10n.apiaryNameRequired;
                            }
                            return validateSizeTier(
                              v,
                              SizeTier.small,
                              l10n.apiaryName,
                              l10n,
                            );
                          },
                        ),
                        const SizedBox(height: 24),
                        ApiaryGridSection(
                          rows: _gridRows,
                          cols: _gridCols,
                          onRowsChanged: (v) => setState(() => _gridRows = v),
                          onColsChanged: (v) => setState(() => _gridCols = v),
                          l10n: l10n,
                        ),
                        if (_gridTooSmall) ...[
                          const SizedBox(height: 8),
                          Text(
                            l10n.apiaryGridTooSmall,
                            style: Theme.of(context).textTheme.bodySmall
                                ?.copyWith(color: Theme.of(context).colorScheme.error),
                          ),
                        ] else if (_outOfBoundsCount > 0) ...[
                          const SizedBox(height: 8),
                          Text(
                            l10n.apiaryGridHivesWillMove(_outOfBoundsCount),
                            style: Theme.of(context).textTheme.bodySmall
                                ?.copyWith(color: Theme.of(context).colorScheme.onSurfaceVariant),
                          ),
                        ],
                        const SizedBox(height: 24),
                        ApiaryLocationSection(
                          latController: _latController,
                          lngController: _lngController,
                          locating: _locating,
                          onGps: () => _useGps(l10n),
                          onMap: _pickOnMap,
                          l10n: l10n,
                        ),
                        const SizedBox(height: 32),
                        Center(
                          child: SizedBox(
                            width: 200,
                            child: ElevatedButton(
                              onPressed: (loading || _gridTooSmall) ? null : () => _submit(l10n),
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

