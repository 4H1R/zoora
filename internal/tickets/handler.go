package tickets

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

// ticketsListConfig white-lists client-controllable search/order columns.
var ticketsListConfig = domain.ListConfig{
	AllowedSearchFields: []string{"title"},
	AllowedOrderFields:  []string{"updated_at", "created_at", "title"},
	DefaultOrderBy:      "updated_at",
	DefaultOrderDir:     "desc",
}

type Handler struct{ svc domain.TicketService }

func NewHandler(svc domain.TicketService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc, perm func(domain.PermissionName) gin.HandlerFunc) {
	id := httpx.RequireUUIDParam("id")

	authed := rg.Group("", authMiddleware)
	{
		authed.GET("/tickets", perm(domain.PermTicketsView), h.List)
		authed.POST("/tickets", perm(domain.PermTicketsView), h.Create)
		authed.GET("/tickets/:id", perm(domain.PermTicketsView), id, h.Get)
		authed.POST("/tickets/:id/messages", perm(domain.PermTicketsView), id, h.AddMessage)
		authed.POST("/tickets/:id/close", perm(domain.PermTicketsView), id, h.Close)
	}
}

// List returns tickets visible to the caller: own tickets plus, for class
// teachers holding tickets:manage, tickets of classes they own.
// @Summary List tickets
// @Description Role-scoped inbox: students see tickets they created; class teachers additionally see tickets of classes they own. Filterable by class_id, status, type. Search matches title. Orderable: updated_at, created_at, title.
// @Tags Tickets
// @Produce json
// @Security BearerAuth
// @Param class_id query string false "Filter by class UUID"
// @Param status query string false "Filter by status: open, answered, closed"
// @Param type query string false "Filter by type: question, grade_objection, other"
// @Param search query string false "Substring match on title"
// @Param order_by query string false "One of: updated_at, created_at, title"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Param page_size query int false "Items per page"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.Ticket}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /tickets [get]
func (h *Handler) List(c *gin.Context) {
	p := listparams.Bind(c, ticketsListConfig)
	q := domain.ListTicketsQuery{ListParams: p}
	if v := c.Query("class_id"); v != "" {
		if cid, err := uuid.Parse(v); err == nil {
			q.ClassID = &cid
		}
	}
	if st := domain.TicketStatus(c.Query("status")); st.Valid() {
		q.Status = &st
	}
	if tt := domain.TicketType(c.Query("type")); tt.Valid() {
		q.Type = &tt
	}
	items, total, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(items, total, p))
}

// Create opens a ticket (with its first message) against a class the caller
// is enrolled in.
// @Summary Create ticket
// @Description Caller must be an enrolled member of class_id. For type=grade_objection, at most one of quiz_room_id / gradebook_column_id may be set (both empty = general objection); targets must belong to the class. media_ids must be presigned via POST /media/presign with model_type=ticket, model_id=<class_id>.
// @Tags Tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body domain.CreateTicketDTO true "Ticket data"
// @Success 201 {object} domain.Response{data=domain.Ticket}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /tickets [post]
func (h *Handler) Create(c *gin.Context) {
	var dto domain.CreateTicketDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	t, err := h.svc.Create(c.Request.Context(), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, t)
}

// Get returns a ticket with its full message thread.
// @Summary Get ticket
// @Description Creator, the class teacher (with tickets:manage), or platform admin. Response includes the full ordered message thread.
// @Tags Tickets
// @Produce json
// @Security BearerAuth
// @Param id path string true "Ticket UUID"
// @Success 200 {object} domain.Response{data=domain.Ticket}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /tickets/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	t, err := h.svc.Get(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, t)
}

// AddMessage appends a reply to an open ticket.
// @Summary Reply to ticket
// @Description Creator or handler only. Handler replies set status=answered; creator replies set status=open. Closed tickets reject replies. Messages are immutable.
// @Tags Tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Ticket UUID"
// @Param body body domain.AddTicketMessageDTO true "Message data"
// @Success 201 {object} domain.Response{data=domain.TicketMessage}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /tickets/{id}/messages [post]
func (h *Handler) AddMessage(c *gin.Context) {
	var dto domain.AddTicketMessageDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	msg, err := h.svc.AddMessage(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, msg)
}

// Close closes a ticket (terminal: no reopen; thread becomes read-only).
// @Summary Close ticket
// @Description Creator or handler. Terminal state — closed tickets accept no further replies and cannot be reopened.
// @Tags Tickets
// @Produce json
// @Security BearerAuth
// @Param id path string true "Ticket UUID"
// @Success 200 {object} domain.Response{data=domain.Ticket}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /tickets/{id}/close [post]
func (h *Handler) Close(c *gin.Context) {
	t, err := h.svc.Close(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, t)
}
