// Package services 的通知範本渲染(07 計畫 Task 4.3.2):純計算、不寫庫。
// 選取鏈:指定語系 → 同 code 同 channel 部門預設語系 → 公司層範本;
// 缺漏變數保留原文 + 警告日誌(不阻斷);渲染結果不再二次解析。
package services

import (
	"context"
	"log"
	"regexp"
	"strings"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/notificationtemplate"
)

// varPattern 為 {{變數}} 佔位(變數名僅英數與底線,全字吻合)。
var varPattern = regexp.MustCompile(`\{\{([A-Za-z0-9_]+)\}\}`)

// RenderTemplate 選取範本並渲染(回 title/content;查無範本回 notFound=true)。
func RenderTemplate(ctx context.Context, db *ent.Client, cid int, did *int, code, channel, locale string, vars map[string]string) (title, content string, notFound bool) {
	tpl := pickTemplate(ctx, db, cid, did, code, channel, locale)
	if tpl == nil {
		return "", "", true
	}
	return renderVars(tpl.Subject, vars), renderVars(tpl.Body, vars), false
}

// pickTemplate 依退回鏈選範本(啟用中 + 未軟刪除)。
func pickTemplate(ctx context.Context, db *ent.Client, cid int, did *int, code, channel, locale string) *ent.NotificationTemplate {
	// 1. 指定語系(部門層)。
	if did != nil {
		if t := queryTemplate(ctx, db, cid, did, code, channel, locale); t != nil {
			return t
		}
		// 2. 同 code 同 channel 部門預設語系。
		if t := queryTemplateAnyLocale(ctx, db, cid, did, code, channel); t != nil {
			return t
		}
	}
	// 3. 公司層範本(department_id IS NULL,先指定語系再任一語系)。
	if t := queryTemplate(ctx, db, cid, nil, code, channel, locale); t != nil {
		return t
	}
	return queryTemplateAnyLocale(ctx, db, cid, nil, code, channel)
}

// queryTemplate 查指定語系範本。
func queryTemplate(ctx context.Context, db *ent.Client, cid int, did *int, code, channel, locale string) *ent.NotificationTemplate {
	q := db.NotificationTemplate.Query().
		Where(notificationtemplate.CompanyIDEQ(cid), notificationtemplate.CodeEQ(code),
			notificationtemplate.ChannelEQ(channel), notificationtemplate.LocaleEQ(locale),
			notificationtemplate.IsActiveEQ(true), notificationtemplate.DeletedAtIsNil())
	if did == nil {
		q = q.Where(notificationtemplate.DepartmentIDIsNil())
	} else {
		q = q.Where(notificationtemplate.DepartmentIDEQ(*did))
	}
	t, err := q.Only(ctx)
	if err != nil {
		return nil
	}
	return t
}

// queryTemplateAnyLocale 查同 code 同 channel 任一語系(訂 id 升冪取首筆,確定性)。
func queryTemplateAnyLocale(ctx context.Context, db *ent.Client, cid int, did *int, code, channel string) *ent.NotificationTemplate {
	q := db.NotificationTemplate.Query().
		Where(notificationtemplate.CompanyIDEQ(cid), notificationtemplate.CodeEQ(code),
			notificationtemplate.ChannelEQ(channel),
			notificationtemplate.IsActiveEQ(true), notificationtemplate.DeletedAtIsNil())
	if did == nil {
		q = q.Where(notificationtemplate.DepartmentIDIsNil())
	} else {
		q = q.Where(notificationtemplate.DepartmentIDEQ(*did))
	}
	t, err := q.Order(ent.Asc(notificationtemplate.FieldID)).First(ctx)
	if err != nil {
		return nil
	}
	return t
}

// renderVars 替換佔位符(缺漏保留原文 + 警告日誌;多餘忽略)。
func renderVars(src string, vars map[string]string) string {
	var missing []string
	out := varPattern.ReplaceAllStringFunc(src, func(m string) string {
		name := varPattern.FindStringSubmatch(m)[1]
		if v, ok := vars[name]; ok {
			return v
		}
		missing = append(missing, name)
		return m
	})
	if len(missing) > 0 {
		log.Printf("notify: 範本缺漏變數 %s(保留原文,不阻斷)", strings.Join(missing, ","))
	}
	return out
}
