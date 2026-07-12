import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/api/api_client.dart';
import '../../../core/theme/app_layout.dart';
import '../../../core/widgets/profile_icon_button.dart';
import '../../../l10n/app_localizations.dart';
import '../../hive/data/hive_model.dart';
import '../data/harvest_model.dart';
import '../data/harvest_repository.dart';
import 'harvest_form_fields.dart';

class HarvestFormScreen extends StatefulWidget {
  final int apiaryId;
  final Hive hive;
  final Harvest? harvest;

  const HarvestFormScreen({
    super.key,
    required this.apiaryId,
    required this.hive,
    this.harvest,
  });

  bool get isEditing => harvest != null;

  @override
  State<HarvestFormScreen> createState() => _HarvestFormScreenState();
}

class _HarvestFormScreenState extends State<HarvestFormScreen> {
  final _formKey = GlobalKey<FormState>();

  late DateTime _harvestedAt;
  late final TextEditingController _framesController;
  late final TextEditingController _halfFramesController;
  late final TextEditingController _kilogramsController;
  late final TextEditingController _notesController;

  bool _loading = false;

  @override
  void initState() {
    super.initState();
    final h = widget.harvest;
    _harvestedAt = h?.harvestedAt ?? DateTime.now();
    _framesController =
        TextEditingController(text: h != null ? '${h.frames}' : '1');
    _halfFramesController =
        TextEditingController(text: h != null ? '${h.halfFrames}' : '0');
    _kilogramsController = TextEditingController(
      text: h != null ? h.kilograms.toStringAsFixed(2) : '',
    );
    _notesController = TextEditingController(text: h?.notes ?? '');
  }

  @override
  void dispose() {
    _framesController.dispose();
    _halfFramesController.dispose();
    _kilogramsController.dispose();
    _notesController.dispose();
    super.dispose();
  }

  Future<void> _pickDate() async {
    final picked = await showDatePicker(
      context: context,
      initialDate: _harvestedAt,
      firstDate: DateTime(2000),
      lastDate: DateTime.now(),
    );
    if (picked == null || !mounted) return;

    final pickedTime = await showTimePicker(
      context: context,
      initialTime: TimeOfDay.fromDateTime(_harvestedAt),
    );
    if (!mounted) return;
    setState(
      () => _harvestedAt = DateTime(
        picked.year,
        picked.month,
        picked.day,
        pickedTime?.hour ?? _harvestedAt.hour,
        pickedTime?.minute ?? _harvestedAt.minute,
      ),
    );
  }

  Future<void> _submit(BuildContext ctx) async {
    if (!(_formKey.currentState?.validate() ?? false)) return;
    setState(() => _loading = true);
    final repo = HarvestRepository(api: ctx.read<ApiClient>());
    final frames = int.tryParse(_framesController.text.trim()) ?? 0;
    final halfFrames = int.tryParse(_halfFramesController.text.trim()) ?? 0;
    final kilograms =
        double.tryParse(_kilogramsController.text.trim().replaceAll(',', '.')) ??
            0.0;
    final notes = _notesController.text.trim();
    try {
      if (widget.isEditing) {
        await repo.updateHarvest(
          apiaryId: widget.apiaryId,
          hiveId: widget.hive.id,
          harvestId: widget.harvest!.id,
          harvestedAt: _harvestedAt,
          frames: frames,
          halfFrames: halfFrames,
          kilograms: kilograms,
          notes: notes,
        );
      } else {
        await repo.createHarvest(
          apiaryId: widget.apiaryId,
          hiveId: widget.hive.id,
          harvestedAt: _harvestedAt,
          frames: frames,
          halfFrames: halfFrames,
          kilograms: kilograms,
          notes: notes,
        );
      }

      if (!ctx.mounted) return;
      Navigator.of(ctx).pop(true);
    } catch (_) {
      if (ctx.mounted) {
        ScaffoldMessenger.of(ctx).showSnackBar(
          SnackBar(content: Text(AppLocalizations.of(ctx)!.generalError)),
        );
      }
      setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      appBar: AppBar(
        title: Text(widget.isEditing ? l10n.harvestEdit : l10n.harvestAdd),
        actions: const [ProfileIconButton()],
      ),
      body: Column(
        children: [
          Expanded(
            child: SafeArea(
              bottom: false,
              child: Center(
                child: SingleChildScrollView(
                  padding: const EdgeInsets.all(24),
                  child: ConstrainedBox(
                    constraints: AppLayout.formConstraints(context),
                    child: HarvestFormFields(
                      formKey: _formKey,
                      harvestedAt: _harvestedAt,
                      framesController: _framesController,
                      halfFramesController: _halfFramesController,
                      kilogramsController: _kilogramsController,
                      notesController: _notesController,
                      onDateTap: _pickDate,
                    ),
                  ),
                ),
              ),
            ),
          ),
          SafeArea(
            top: false,
            child: Padding(
              padding: const EdgeInsets.only(bottom: 8),
              child: Center(
                child: SizedBox(
                  width: AppLayout.bannerWidth(context),
                  child: Container(
                    decoration: const BoxDecoration(
                      color: Colors.amber,
                      borderRadius: BorderRadius.only(
                        topLeft: Radius.circular(20),
                        topRight: Radius.circular(20),
                      ),
                      boxShadow: [
                        BoxShadow(
                          color: Colors.black26,
                          blurRadius: 8,
                          offset: Offset(0, -2),
                        ),
                      ],
                    ),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                      children: [
                        _loading
                            ? const Padding(
                                padding: EdgeInsets.all(12),
                                child: SizedBox(
                                  width: 24,
                                  height: 24,
                                  child: CircularProgressIndicator(
                                      strokeWidth: 2),
                                ),
                              )
                            : IconButton(
                                icon: const Icon(Icons.check),
                                iconSize: 28,
                                tooltip: l10n.generalSave,
                                onPressed: () => _submit(context),
                              ),
                      ],
                    ),
                  ),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}
