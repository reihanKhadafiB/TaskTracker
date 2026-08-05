package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-pdf/fpdf"
	"github.com/username/task-tracker/internal/model"
	"github.com/username/task-tracker/internal/service"
)

type ExportHandler struct {
	contextService *service.ContextService
	taskService    *service.TaskService
	subtaskService *service.SubtaskService
}

func NewExportHandler(
	contextService *service.ContextService,
	taskService *service.TaskService,
	subtaskService *service.SubtaskService,
) *ExportHandler {
	return &ExportHandler{
		contextService: contextService,
		taskService:    taskService,
		subtaskService: subtaskService,
	}
}

func (h *ExportHandler) ExportContextPDF(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid context id")
		return
	}

	contexts, err := h.contextService.ListContexts(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get contexts")
		return
	}

	var targetContext *model.Context
	for i := range contexts {
		if contexts[i].ID == id {
			targetContext = &contexts[i]
			break
		}
	}

	if targetContext == nil {
		respondError(w, http.StatusNotFound, "context not found")
		return
	}

	tasks, err := h.taskService.ListTasks(r.Context(), model.TaskFilter{
		ContextID: &id,
		Limit:     1000,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch tasks")
		return
	}

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 20, 15)
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 22)
	pdf.SetTextColor(33, 37, 41)
	pdf.CellFormat(100, 10, "Task Tracker", "", 0, "L", false, 0, "")

	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(108, 117, 125)

	nowStr := time.Now().Format("02 Jan 2006 15:04:05")
	pdf.CellFormat(80, 10, fmt.Sprintf("Date: %s", time.Now().Format("02 Jan 2006")), "", 1, "R", false, 0, "")
	pdf.SetX(15)
	pdf.CellFormat(180, 5, fmt.Sprintf("Generated on: %s", nowStr), "", 1, "R", false, 0, "")

	pdf.Ln(8)

	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(73, 80, 87)
	pdf.CellFormat(180, 8, fmt.Sprintf("Context Report: %s", targetContext.Name), "", 1, "L", false, 0, "")

	pdf.SetDrawColor(222, 226, 230)
	pdf.SetLineWidth(0.5)
	pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
	pdf.Ln(8)

	if len(tasks) == 0 {
		pdf.SetTextColor(108, 117, 125)
		pdf.SetFont("Arial", "I", 11)
		pdf.Cell(40, 10, "No tasks available in this context.")
	} else {
		pdf.SetFillColor(241, 243, 245)
		pdf.SetTextColor(73, 80, 87)
		pdf.SetFont("Arial", "B", 11)
		pdf.CellFormat(15, 10, "Status", "B", 0, "C", true, 0, "")
		pdf.CellFormat(130, 10, "Task / Subtask Name", "B", 0, "L", true, 0, "")
		pdf.CellFormat(35, 10, "Due Date", "B", 1, "C", true, 0, "")

		pdf.SetDrawColor(222, 226, 230)

		for _, t := range tasks {
			pdf.SetTextColor(33, 37, 41)
			pdf.SetFont("Arial", "B", 11)

			dueDateStr := "-"
			if t.DueDate != nil {
				dueDateStr = t.DueDate.Time.Format("02/01/2006")
			}

			pdf.SetFillColor(255, 255, 255)

			x := pdf.GetX()
			y := pdf.GetY()

			pdf.CellFormat(15, 10, "", "B", 0, "C", true, 0, "")
			pdf.CellFormat(130, 10, " "+t.Title, "B", 0, "L", true, 0, "")
			pdf.CellFormat(35, 10, dueDateStr, "B", 1, "C", true, 0, "")

			pdf.SetDrawColor(100, 100, 100)
			pdf.Rect(x+5.5, y+3, 4, 4, "D")
			if t.Status == "done" {
				pdf.SetDrawColor(40, 167, 69)
				pdf.SetLineWidth(0.6)
				pdf.Line(x+6, y+5, x+7, y+6.5)
				pdf.Line(x+7, y+6.5, x+9, y+3.5)
				pdf.SetLineWidth(0.2)
			}
			pdf.SetDrawColor(222, 226, 230)

			subtasks, err := h.subtaskService.ListSubtasks(r.Context(), t.ID)
			if err == nil && len(subtasks) > 0 {
				pdf.SetFont("Arial", "", 10)
				for _, st := range subtasks {
					pdf.SetTextColor(33, 37, 41)

					sx := pdf.GetX()
					sy := pdf.GetY()

					pdf.CellFormat(15, 8, "", "B", 0, "C", true, 0, "")
					pdf.CellFormat(10, 8, "", "B", 0, "R", true, 0, "")
					pdf.CellFormat(120, 8, " "+st.Title, "B", 0, "L", true, 0, "")
					pdf.CellFormat(35, 8, "", "B", 1, "C", true, 0, "")

					pdf.SetDrawColor(100, 100, 100)
					pdf.Rect(sx+19, sy+2, 4, 4, "D")
					if st.IsDone {
						pdf.SetDrawColor(40, 167, 69)
						pdf.SetLineWidth(0.6)
						pdf.Line(sx+19.5, sy+4, sx+20.5, sy+5.5)
						pdf.Line(sx+20.5, sy+5.5, sx+22.5, sy+2.5)
						pdf.SetLineWidth(0.2)
					}
					pdf.SetDrawColor(222, 226, 230)
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"Context_%s_Report.pdf\"", targetContext.Name))

	if err := pdf.Output(w); err != nil {
		http.Error(w, "failed to generate pdf", http.StatusInternalServerError)
	}
}
