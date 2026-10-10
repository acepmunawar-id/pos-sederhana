package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/posumkm/backend/internal/domain"
	"github.com/posumkm/backend/internal/dto"
	"github.com/posumkm/backend/internal/service"
	"github.com/posumkm/backend/pkg/response"
)

type TransactionHandler struct {
	svc *service.TransactionService
	loc *time.Location
}

func NewTransactionHandler(svc *service.TransactionService, loc *time.Location) *TransactionHandler {
	return &TransactionHandler{svc: svc, loc: loc}
}

// Checkout godoc
// @Summary  Checkout (buat transaksi penjualan)
// @Description Atomic: simpan transaksi, item, pembayaran, kurangi stok, catat pergerakan stok.
// @Description Kirim header Idempotency-Key unik per transaksi (mis. UUID). Request ulang dengan key yang sama
// @Description mengembalikan transaksi yang sama (HTTP 200) tanpa membuat transaksi baru.
// @Description discount_value untuk PERCENT dalam basis point (1000 = 10%).
// @Tags     Transactions
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    Idempotency-Key header string true "Kunci unik transaksi (8–100 karakter)"
// @Param    body body dto.CheckoutRequest true "Keranjang & pembayaran"
// @Success  201 {object} response.Body{data=dto.TransactionResponse} "Transaksi baru"
// @Success  200 {object} response.Body{data=dto.TransactionResponse} "Replay idempotent"
// @Failure  422 {object} response.ErrorBody "Stok tidak cukup / pembayaran kurang / belum buka shift"
// @Router   /transactions/checkout [post]
func (h *TransactionHandler) Checkout(c *gin.Context) {
	var req dto.CheckoutRequest
	if !bindJSON(c, &req) {
		return
	}
	a := actor(c)
	t, created, err := h.svc.Checkout(c.Request.Context(), a, c.GetHeader("Idempotency-Key"), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	res := dto.ToTransactionResponse(t, a.IsManager())
	if created {
		response.Created(c, "Transaksi berhasil", res)
		return
	}
	c.Header("Idempotent-Replayed", "true")
	response.OK(c, "Transaksi sudah tercatat sebelumnya", res)
}

// List godoc
// @Summary  Daftar transaksi
// @Description Kasir hanya melihat transaksinya sendiri.
// @Tags     Transactions
// @Security BearerAuth
// @Produce  json
// @Param    page query int false "Halaman" default(1)
// @Param    limit query int false "Jumlah per halaman" default(10)
// @Param    search query string false "Cari nomor invoice / nama kasir"
// @Param    sort query string false "created_at|total|invoice_number|cashier"
// @Param    order query string false "asc|desc"
// @Param    status query string false "COMPLETED|VOIDED"
// @Param    user_id query int false "Filter kasir"
// @Param    shift_id query int false "Filter shift"
// @Param    method query string false "CASH|QRIS|TRANSFER|DEBIT"
// @Param    date_from query string false "YYYY-MM-DD"
// @Param    date_to query string false "YYYY-MM-DD"
// @Success  200 {object} response.Body{data=[]dto.TransactionResponse}
// @Router   /transactions [get]
func (h *TransactionHandler) List(c *gin.Context) {
	f := domain.TransactionFilter{Params: listParams(c)}
	var err error
	if f.Status, err = queryEnum(c, "status", domain.TxCompleted, domain.TxVoided); err != nil {
		response.Error(c, err)
		return
	}
	if f.Method, err = queryEnum(c, "method", domain.PaymentCash, domain.PaymentQRIS, domain.PaymentTransfer, domain.PaymentDebit); err != nil {
		response.Error(c, err)
		return
	}
	if f.UserID, err = queryInt64(c, "user_id"); err != nil {
		response.Error(c, err)
		return
	}
	if f.ShiftID, err = queryInt64(c, "shift_id"); err != nil {
		response.Error(c, err)
		return
	}
	if f.DateFrom, f.DateTo, err = dateFilter(c, h.loc); err != nil {
		response.Error(c, err)
		return
	}
	a := actor(c)
	items, total, err := h.svc.List(c.Request.Context(), a, f)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, "Berhasil", dto.ToTransactionResponses(items, a.IsManager()), f.Meta(total))
}

// Get godoc
// @Summary  Detail transaksi
// @Tags     Transactions
// @Security BearerAuth
// @Produce  json
// @Param    id path int true "ID transaksi"
// @Success  200 {object} response.Body{data=dto.TransactionResponse}
// @Router   /transactions/{id} [get]
func (h *TransactionHandler) Get(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	a := actor(c)
	t, err := h.svc.Get(c.Request.Context(), a, id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Berhasil", dto.ToTransactionResponse(t, a.IsManager()))
}

// Receipt godoc
// @Summary  Data struk
// @Description Data toko + transaksi + teks struk 32 kolom yang siap dibagikan / dicetak.
// @Tags     Transactions
// @Security BearerAuth
// @Produce  json
// @Param    id path int true "ID transaksi"
// @Success  200 {object} response.Body{data=dto.ReceiptResponse}
// @Router   /transactions/{id}/receipt [get]
func (h *TransactionHandler) Receipt(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	t, st, text, err := h.svc.Receipt(c.Request.Context(), actor(c), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Berhasil", dto.ReceiptResponse{
		Store: dto.ToStoreSettingResponse(st), Transaction: dto.ToTransactionResponse(t, false), Text: text,
	})
}

// Void godoc
// @Summary  Void transaksi
// @Description Hanya Owner/Admin. Stok semua item otomatis dikembalikan.
// @Tags     Transactions
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    id path int true "ID transaksi"
// @Param    body body dto.VoidRequest true "Alasan pembatalan"
// @Success  200 {object} response.Body{data=dto.TransactionResponse}
// @Failure  409 {object} response.ErrorBody "Sudah di-void"
// @Router   /transactions/{id}/void [post]
func (h *TransactionHandler) Void(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	var req dto.VoidRequest
	if !bindJSON(c, &req) {
		return
	}
	t, err := h.svc.Void(c.Request.Context(), actor(c), id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Transaksi berhasil dibatalkan dan stok dikembalikan", dto.ToTransactionResponse(t, true))
}
