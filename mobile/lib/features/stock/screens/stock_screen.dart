import 'dart:async';

import 'package:flutter/cupertino.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/api_exception.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/app_button.dart';
import '../../../core/widgets/app_scaffold.dart';
import '../../../core/widgets/feedback.dart';
import '../../../core/widgets/inputs.dart';
import '../../../core/widgets/product_image.dart';
import '../../auth/providers/auth_controller.dart';
import '../../catalog/models/product.dart';
import '../../catalog/providers/catalog_providers.dart';
import '../data/stock_repository.dart';
import '../providers/stock_providers.dart';

class StockScreen extends ConsumerStatefulWidget {
  const StockScreen({super.key});

  @override
  ConsumerState<StockScreen> createState() => _StockScreenState();
}

class _StockScreenState extends ConsumerState<StockScreen> {
  StockQuery _query = const StockQuery();
  Timer? _debounce;

  @override
  void dispose() {
    _debounce?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    final list = ref.watch(stockListProvider(_query));
    final canManage = ref.watch(currentUserProvider).isManager;

    return LargeTitleScaffold(
      title: 'Stok',
      onRefresh: () async {
        ref.invalidate(stockListProvider(_query));
        await ref.read(stockListProvider(_query).future);
      },
      slivers: [
        SliverToBoxAdapter(
          child: Padding(
            padding: const EdgeInsets.fromLTRB(AppSpacing.lg, AppSpacing.sm, AppSpacing.lg, 0),
            child: Column(children: [
              CupertinoSearchTextField(
                placeholder: 'Cari produk',
                backgroundColor: c.fill,
                style: AppText.body.copyWith(color: c.text),
                onChanged: (v) {
                  _debounce?.cancel();
                  _debounce = Timer(const Duration(milliseconds: 300), () => setState(() => _query = StockQuery(search: v.trim(), lowOnly: _query.lowOnly)));
                },
              ),
              const SizedBox(height: AppSpacing.md),
              AppSegmented<bool>(
                value: _query.lowOnly,
                segments: const {false: 'Semua', true: 'Stok Menipis'},
                onChanged: (v) => setState(() => _query = StockQuery(search: _query.search, lowOnly: v)),
              ),
            ]),
          ),
        ),
        ...list.when(
          loading: () => [
            SliverList.builder(
              itemCount: 8,
              itemBuilder: (_, __) => const Padding(
                padding: EdgeInsets.fromLTRB(AppSpacing.lg, AppSpacing.md, AppSpacing.lg, 0),
                child: Skeleton(height: 60, radius: AppRadius.sm),
              ),
            ),
          ],
          error: (e, _) => [
            SliverToBoxAdapter(child: ErrorState(message: errorMessage(e), onRetry: () => ref.invalidate(stockListProvider(_query)))),
          ],
          data: (items) => items.isEmpty
              ? [
                  SliverToBoxAdapter(
                    child: EmptyState(
                      icon: _query.lowOnly ? CupertinoIcons.checkmark_seal : CupertinoIcons.cube_box,
                      title: _query.lowOnly ? 'Semua stok aman' : 'Produk tidak ditemukan',
                      message: _query.lowOnly ? 'Tidak ada produk di bawah stok minimum.' : null,
                    ),
                  ),
                ]
              : [
                  SliverPadding(
                    padding: const EdgeInsets.fromLTRB(AppSpacing.lg, AppSpacing.lg, AppSpacing.lg, 0),
                    sliver: SliverToBoxAdapter(
                      child: ClipRRect(
                        borderRadius: BorderRadius.circular(AppRadius.sm),
                        child: ColoredBox(
                          color: c.card,
                          child: Column(children: [
                            for (final (i, p) in items.indexed) ...[
                              if (i > 0) Padding(padding: const EdgeInsets.only(left: 76), child: Container(height: 0.5, color: c.separator)),
                              _StockRow(product: p, onTap: canManage ? () => _openActions(p) : null),
                            ],
                          ]),
                        ),
                      ),
                    ),
                  ),
                  if (canManage)
                    SliverToBoxAdapter(
                      child: Padding(
                        padding: const EdgeInsets.fromLTRB(AppSpacing.xl, AppSpacing.md, AppSpacing.xl, 0),
                        child: Text('Ketuk produk untuk menambah stok atau menyesuaikan hasil hitung.',
                            style: AppText.footnote.copyWith(color: c.textSecondary)),
                      ),
                    ),
                ],
        ),
      ],
    );
  }

  Future<void> _openActions(Product p) async {
    final changed = await showAppSheet<bool>(context, title: p.name, child: _StockForm(product: p));
    if (changed == true) {
      ref.invalidate(stockListProvider);
      ref.invalidate(productListProvider);
    }
  }
}

class _StockRow extends StatelessWidget {
  const _StockRow({required this.product, this.onTap});
  final Product product;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    final color = product.isOutOfStock ? c.danger : product.isLowStock ? c.warning : c.text;
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        child: Padding(
          padding: const EdgeInsets.all(AppSpacing.md),
          child: Row(children: [
            ProductImage(url: product.imageUrl, size: 48),
            const SizedBox(width: AppSpacing.md),
            Expanded(
              child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                Text(product.name, style: AppText.body.copyWith(color: c.text, fontWeight: FontWeight.w500)),
                Text('${product.categoryName} · min ${product.minStock}', style: AppText.footnote.copyWith(color: c.textSecondary)),
              ]),
            ),
            Column(crossAxisAlignment: CrossAxisAlignment.end, children: [
              Text('${product.stock}', style: AppText.title.copyWith(color: color, fontSize: 22)),
              Text(
                product.isOutOfStock ? 'Habis' : product.isLowStock ? 'Menipis' : product.unitName,
                style: AppText.caption.copyWith(color: product.isLowStock ? color : c.textSecondary, fontWeight: FontWeight.w600),
              ),
            ]),
          ]),
        ),
      ),
    );
  }
}

enum _StockAction { stockIn, adjust }

/// Stock-in / adjustment form (Owner & Admin only; enforced by the API too).
class _StockForm extends ConsumerStatefulWidget {
  const _StockForm({required this.product});
  final Product product;

  @override
  ConsumerState<_StockForm> createState() => _StockFormState();
}

class _StockFormState extends ConsumerState<_StockForm> {
  _StockAction _action = _StockAction.stockIn;
  final _qty = TextEditingController();
  final _note = TextEditingController();
  bool _saving = false;
  String? _error;

  @override
  void dispose() {
    _qty.dispose();
    _note.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    final n = int.tryParse(_qty.text) ?? -1;
    final note = _note.text.trim();
    if (_action == _StockAction.stockIn && n <= 0) return setState(() => _error = 'Masukkan jumlah barang masuk');
    if (_action == _StockAction.adjust && n < 0) return setState(() => _error = 'Masukkan stok hasil hitung');
    if (_action == _StockAction.adjust && note.length < 3) return setState(() => _error = 'Tulis alasan penyesuaian');
    setState(() {
      _error = null;
      _saving = true;
    });
    try {
      final repo = ref.read(stockRepositoryProvider);
      if (_action == _StockAction.stockIn) {
        await repo.stockIn(widget.product.id, n, note: note);
      } else {
        await repo.adjust(widget.product.id, n, note);
      }
      if (!mounted) return;
      Navigator.pop(context, true);
      showToast(context, _action == _StockAction.stockIn ? 'Stok ${widget.product.name} bertambah $n' : 'Stok disesuaikan menjadi $n');
    } catch (e) {
      if (mounted) setState(() => _error = errorMessage(e));
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    final isIn = _action == _StockAction.stockIn;
    return Column(crossAxisAlignment: CrossAxisAlignment.stretch, children: [
      Text('Stok sekarang: ${widget.product.stock} ${widget.product.unitName}',
          textAlign: TextAlign.center, style: AppText.subheadline.copyWith(color: c.textSecondary)),
      const SizedBox(height: AppSpacing.lg),
      AppSegmented<_StockAction>(
        value: _action,
        segments: const {_StockAction.stockIn: 'Stok Masuk', _StockAction.adjust: 'Penyesuaian'},
        onChanged: (a) => setState(() {
          _action = a;
          _error = null;
          _qty.clear();
        }),
      ),
      const SizedBox(height: AppSpacing.lg),
      AppTextField(
        key: ValueKey(_action),
        label: isIn ? 'Jumlah barang masuk' : 'Stok hasil hitung fisik',
        controller: _qty,
        placeholder: '0',
        large: true,
        autofocus: true,
        keyboardType: TextInputType.number,
        inputFormatters: [FilteringTextInputFormatter.digitsOnly, LengthLimitingTextInputFormatter(7)],
        error: _error,
      ),
      const SizedBox(height: AppSpacing.md),
      AppTextField(
        label: isIn ? 'Catatan (opsional)' : 'Alasan',
        controller: _note,
        placeholder: isIn ? 'Contoh: dari Supplier A' : 'Contoh: barang rusak',
      ),
      const SizedBox(height: AppSpacing.xl),
      AppButton(label: 'Simpan', icon: CupertinoIcons.checkmark_alt, loading: _saving, onPressed: _save),
    ]);
  }
}
