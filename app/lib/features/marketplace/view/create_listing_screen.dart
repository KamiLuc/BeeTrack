import 'dart:typed_data';

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:image_picker/image_picker.dart';

import '../../../core/api/api_client.dart';
import '../../../core/theme/app_layout.dart';
import '../../../core/widgets/profile_icon_button.dart';
import '../../../features/apiary/data/apiary_model.dart';
import '../../../features/apiary/data/apiary_repository.dart';
import '../../../l10n/app_localizations.dart';
import '../data/listing_category.dart';
import '../data/listing_repository.dart';

const int _kMaxListingPhotos = 3;

class CreateListingScreen extends StatefulWidget {
  const CreateListingScreen({super.key});

  @override
  State<CreateListingScreen> createState() => _CreateListingScreenState();
}

class _CreateListingScreenState extends State<CreateListingScreen> {
  final _formKey = GlobalKey<FormState>();
  final _titleController = TextEditingController();
  final _descriptionController = TextEditingController();
  final _priceController = TextEditingController();
  final _quantityController = TextEditingController();
  final _addressController = TextEditingController();
  final _phoneController = TextEditingController();
  final _emailController = TextEditingController();

  String? _category;
  int? _apiaryId;
  List<Apiary> _apiaries = [];
  final List<XFile> _pendingImages = [];
  bool _saving = false;
  String? _submitError;

  @override
  void initState() {
    super.initState();
    _loadApiaries();
  }

  @override
  void dispose() {
    _titleController.dispose();
    _descriptionController.dispose();
    _priceController.dispose();
    _quantityController.dispose();
    _addressController.dispose();
    _phoneController.dispose();
    _emailController.dispose();
    super.dispose();
  }

  Future<void> _loadApiaries() async {
    try {
      final apiaries =
          await ApiaryRepository(api: context.read<ApiClient>()).listApiaries();
      if (mounted) setState(() => _apiaries = apiaries);
    } catch (_) {}
  }

  bool _isValidEmail(String v) {
    return RegExp(r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$')
        .hasMatch(v.trim());
  }

  bool _isValidPhone(String v) {
    final digits = v.replaceAll(RegExp(r'[^0-9]'), '');
    return digits.length >= 9;
  }

  static final _priceInputFormatter = TextInputFormatter.withFunction(
    (oldValue, newValue) {
      if (newValue.text.isEmpty) return newValue;
      return RegExp(r'^\d*\.?\d{0,2}$').hasMatch(newValue.text)
          ? newValue
          : oldValue;
    },
  );

  Future<void> _pickImage() async {
    if (_pendingImages.length >= _kMaxListingPhotos) return;
    final l10n = AppLocalizations.of(context)!;
    final picker = ImagePicker();
    final ImageSource source;
    if (kIsWeb) {
      source = ImageSource.gallery;
    } else {
      final chosen = await showModalBottomSheet<ImageSource>(
        context: context,
        builder: (_) => SafeArea(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              ListTile(
                leading: const Icon(Icons.photo_library_outlined),
                title: Text(l10n.marketplacePhotoSourceGallery),
                onTap: () => Navigator.of(context).pop(ImageSource.gallery),
              ),
              ListTile(
                leading: const Icon(Icons.camera_alt_outlined),
                title: Text(l10n.marketplacePhotoSourceCamera),
                onTap: () => Navigator.of(context).pop(ImageSource.camera),
              ),
            ],
          ),
        ),
      );
      if (chosen == null) return;
      source = chosen;
    }
    final file = await picker.pickImage(source: source, imageQuality: 85);
    if (file != null && mounted) setState(() => _pendingImages.add(file));
  }

  void _removeImage(XFile file) {
    setState(() => _pendingImages.remove(file));
  }

  void _reviewContactError() {
    if (_submitError == null) return;
    if (_submitError != AppLocalizations.of(context)!.marketplaceContactRequired) {
      return;
    }
    final hasContact = _phoneController.text.trim().isNotEmpty ||
        _emailController.text.trim().isNotEmpty;
    if (hasContact) setState(() => _submitError = null);
  }

  Future<void> _submit() async {
    final l10n = AppLocalizations.of(context)!;
    final formValid = _formKey.currentState?.validate() ?? false;
    final hasContact = _phoneController.text.trim().isNotEmpty ||
        _emailController.text.trim().isNotEmpty;
    if (!formValid || _category == null || !hasContact) {
      setState(() {
        _submitError = !hasContact ? l10n.marketplaceContactRequired : null;
      });
      return;
    }

    setState(() {
      _submitError = null;
      _saving = true;
    });
    final repo = ListingRepository(api: context.read<ApiClient>());
    try {
      final listing = await repo.createListing(
        title: _titleController.text.trim(),
        description: _descriptionController.text.trim(),
        category: _category!,
        price: double.parse(_priceController.text.trim()),
        quantity: _quantityController.text.trim(),
        address: _addressController.text.trim(),
        apiaryId: _apiaryId,
        contactPhone: _phoneController.text.trim(),
        contactEmail: _emailController.text.trim(),
      );
      for (final file in _pendingImages) {
        await repo.uploadImage(listing.id, file);
      }
      if (mounted) Navigator.of(context).pop(true);
    } catch (_) {
      if (mounted) {
        setState(() {
          _submitError = l10n.generalError;
          _saving = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final textTheme = Theme.of(context).textTheme;
    final colorScheme = Theme.of(context).colorScheme;

    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.marketplaceCreateScreenTitle),
        actions: const [ProfileIconButton()],
      ),
      body: Center(
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
                    controller: _titleController,
                    decoration:
                        InputDecoration(labelText: l10n.marketplaceFieldTitle),
                    textInputAction: TextInputAction.next,
                    autovalidateMode: AutovalidateMode.onUserInteraction,
                    validator: (v) => (v == null || v.trim().isEmpty)
                        ? l10n.marketplaceFieldTitleRequired
                        : null,
                  ),
                  const SizedBox(height: 16),
                  DropdownButtonFormField<String>(
                    initialValue: _category,
                    decoration:
                        InputDecoration(labelText: l10n.marketplaceFieldCategory),
                    autovalidateMode: AutovalidateMode.onUserInteraction,
                    items: [
                      for (final category in listingCategories)
                        DropdownMenuItem(
                          value: category,
                          child: Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              Icon(listingCategoryIcon(category), size: 20),
                              const SizedBox(width: 8),
                              Text(listingCategoryLabel(l10n, category)),
                            ],
                          ),
                        ),
                    ],
                    onChanged: (value) => setState(() => _category = value),
                    validator: (v) =>
                        v == null ? l10n.marketplaceFieldCategoryRequired : null,
                  ),
                  const SizedBox(height: 16),
                  TextFormField(
                    controller: _descriptionController,
                    decoration: InputDecoration(
                      labelText: l10n.marketplaceDescriptionLabel,
                    ),
                    minLines: 3,
                    maxLines: 6,
                    maxLength: 500,
                    textInputAction: TextInputAction.newline,
                  ),
                  const SizedBox(height: 16),
                  TextFormField(
                    controller: _priceController,
                    decoration: InputDecoration(
                      labelText: l10n.marketplaceFieldPrice,
                    ),
                    keyboardType:
                        TextInputType.numberWithOptions(decimal: true),
                    textInputAction: TextInputAction.next,
                    inputFormatters: [_priceInputFormatter],
                    autovalidateMode: AutovalidateMode.onUserInteraction,
                    validator: (v) {
                      if (v == null || v.trim().isEmpty) {
                        return l10n.marketplaceFieldPriceRequired;
                      }
                      return double.tryParse(v.trim()) == null
                          ? l10n.marketplaceFieldPriceInvalid
                          : null;
                    },
                  ),
                  const SizedBox(height: 16),
                  TextFormField(
                    controller: _quantityController,
                    decoration: InputDecoration(
                      labelText: l10n.marketplaceQuantityLabel,
                    ),
                    textInputAction: TextInputAction.next,
                  ),
                  const SizedBox(height: 16),
                  TextFormField(
                    controller: _addressController,
                    decoration: InputDecoration(
                      labelText: l10n.marketplaceFieldAddress,
                    ),
                    textInputAction: TextInputAction.next,
                  ),
                  const SizedBox(height: 24),
                  Text(
                    l10n.marketplaceContactLabel,
                    style: textTheme.titleSmall
                        ?.copyWith(color: colorScheme.primary),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    l10n.marketplaceContactRequired,
                    style: textTheme.bodySmall
                        ?.copyWith(color: colorScheme.onSurfaceVariant),
                  ),
                  const SizedBox(height: 8),
                  TextFormField(
                    controller: _phoneController,
                    decoration:
                        InputDecoration(labelText: l10n.marketplaceFieldPhone),
                    keyboardType: TextInputType.phone,
                    textInputAction: TextInputAction.next,
                    autovalidateMode: AutovalidateMode.onUserInteraction,
                    onChanged: (_) => _reviewContactError(),
                    validator: (v) {
                      if (v == null || v.trim().isEmpty) return null;
                      return _isValidPhone(v)
                          ? null
                          : l10n.marketplaceFieldPhoneInvalid;
                    },
                  ),
                  const SizedBox(height: 16),
                  TextFormField(
                    controller: _emailController,
                    decoration:
                        InputDecoration(labelText: l10n.marketplaceFieldEmail),
                    keyboardType: TextInputType.emailAddress,
                    textInputAction: TextInputAction.next,
                    autovalidateMode: AutovalidateMode.onUserInteraction,
                    onChanged: (_) => _reviewContactError(),
                    validator: (v) {
                      if (v == null || v.trim().isEmpty) return null;
                      return _isValidEmail(v) ? null : l10n.authInvalidEmail;
                    },
                  ),
                  if (_apiaries.isNotEmpty) ...[
                    const SizedBox(height: 24),
                    Text(
                      l10n.marketplaceApiaryLabel,
                      style: textTheme.titleSmall
                          ?.copyWith(color: colorScheme.primary),
                    ),
                    const SizedBox(height: 8),
                    DropdownButtonFormField<int?>(
                      initialValue: _apiaryId,
                      items: [
                        DropdownMenuItem(
                          value: null,
                          child: Text(l10n.marketplaceApiaryNone),
                        ),
                        for (final apiary in _apiaries)
                          DropdownMenuItem(
                            value: apiary.id,
                            child: Text(apiary.name),
                          ),
                      ],
                      onChanged: (value) => setState(() => _apiaryId = value),
                    ),
                  ],
                  const SizedBox(height: 24),
                  _PhotosSection(
                    pendingImages: _pendingImages,
                    onAdd: _pickImage,
                    onRemove: _removeImage,
                  ),
                  const SizedBox(height: 24),
                  if (_submitError != null) ...[
                    Text(
                      _submitError!,
                      textAlign: TextAlign.center,
                      style: TextStyle(color: colorScheme.error),
                    ),
                    const SizedBox(height: 12),
                  ],
                  Center(
                    child: SizedBox(
                      width: 200,
                      child: ElevatedButton(
                        onPressed: _saving ? null : _submit,
                        child: _saving
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
      ),
    );
  }
}

class _PhotosSection extends StatelessWidget {
  final List<XFile> pendingImages;
  final VoidCallback onAdd;
  final ValueChanged<XFile> onRemove;

  const _PhotosSection({
    required this.pendingImages,
    required this.onAdd,
    required this.onRemove,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final atLimit = pendingImages.length >= _kMaxListingPhotos;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(
              '${l10n.marketplacePhotosLabel}  ${pendingImages.length}/$_kMaxListingPhotos',
              style: Theme.of(context).textTheme.titleSmall?.copyWith(
                    color: Theme.of(context).colorScheme.primary,
                  ),
            ),
            TextButton.icon(
              onPressed: atLimit ? null : onAdd,
              icon: const Icon(Icons.add_photo_alternate_outlined, size: 18),
              label: Text(l10n.marketplaceAddPhoto),
            ),
          ],
        ),
        if (pendingImages.isNotEmpty)
          SizedBox(
            height: 100,
            child: ListView(
              shrinkWrap: true,
              scrollDirection: Axis.horizontal,
              children: [
                for (final file in pendingImages)
                  _PendingThumbnail(file: file, onRemove: () => onRemove(file)),
              ],
            ),
          ),
      ],
    );
  }
}

class _PendingThumbnail extends StatelessWidget {
  final XFile file;
  final VoidCallback onRemove;

  const _PendingThumbnail({required this.file, required this.onRemove});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(right: 8, bottom: 4),
      child: Stack(
        children: [
          ClipRRect(
            borderRadius: BorderRadius.circular(8),
            child: SizedBox(
              width: 90,
              height: 90,
              child: FutureBuilder<Uint8List>(
                future: file.readAsBytes(),
                builder: (_, snapshot) {
                  if (snapshot.hasData) {
                    return Image.memory(snapshot.data!, fit: BoxFit.cover);
                  }
                  return const Center(
                    child: CircularProgressIndicator(strokeWidth: 2),
                  );
                },
              ),
            ),
          ),
          Positioned(
            top: 2,
            right: 2,
            child: GestureDetector(
              onTap: onRemove,
              child: Container(
                decoration: const BoxDecoration(
                  color: Colors.black54,
                  shape: BoxShape.circle,
                ),
                child: const Icon(Icons.close, color: Colors.white, size: 18),
              ),
            ),
          ),
        ],
      ),
    );
  }
}
