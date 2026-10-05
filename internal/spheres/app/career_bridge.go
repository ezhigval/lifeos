package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

// CareerBridge — мост «Карьера → воркспейс» (TASK-010 WS-15 правило 3,
// TASK-011 п.8): сфера «Карьера» может быть привязана к внешней доменной
// сущности через sphere_domain_links(link_type='workspace'). Пока воркспейсов
// в системе нет (TASK-010 Stage 4 не запущен), в качестве цели моста
// используется contact из career_contacts: он даёт отображаемое имя и id.
type DomainLinkReader interface {
	DomainLink(ctx context.Context, userID ids.UserID, sphereID ids.SphereID, linkType string) (refID ids.ContactID, refName string, found bool, err error)
}

// CareerContactLister читает имена контактов Карьеры (реализует infra-обёртка над career_contacts).
type CareerContactLister interface {
	ListContactNames(ctx context.Context, userID ids.UserID) ([]string, error)
}

type SphereDomainLinkDTO struct {
	LinkType string
	RefID    ids.ContactID
	RefName  string
}

type GetSphereDomainLink struct {
	links       DomainLinkReader
	spheres     SphereStore
	contacts    CareerContactLister // опционально, may be nil
	careerNames map[string]bool
}

func NewGetSphereDomainLink(links DomainLinkReader, spheres SphereStore, contacts CareerContactLister) *GetSphereDomainLink {
	return &GetSphereDomainLink{
		links:    links,
		spheres:  spheres,
		contacts: contacts,
		careerNames: map[string]bool{
			strings.ToLower("Карьера"):  true,
			strings.ToLower("Career"):   true,
			strings.ToLower("Работа"):   true,
			strings.ToLower("Работа 💼"): true,
		},
	}
}

// SphereIsCareer определяет, является ли сфера карьерной: по имени или
// по наличию любого domain-линка (пользователь мог связать произвольную сферу вручную).
func (uc *GetSphereDomainLink) SphereIsCareer(ctx context.Context, userID ids.UserID, name string) bool {
	if uc.careerNames[strings.ToLower(strings.TrimSpace(name))] {
		return true
	}
	sphere, err := uc.spheres.FindByName(ctx, userID, name)
	if err != nil {
		return false
	}
	_, _, found, err := uc.links.DomainLink(ctx, userID, sphere.ID, "workspace")
	return err == nil && found
}

// Execute возвращает workspace-линк сферы (found=false, если не привязан).
func (uc *GetSphereDomainLink) Execute(ctx context.Context, userID ids.UserID, sphereID ids.SphereID) (SphereDomainLinkDTO, bool, error) {
	if userID.IsZero() || sphereID.IsZero() {
		return SphereDomainLinkDTO{}, false, fmt.Errorf("user id and sphere id are required")
	}
	refID, refName, found, err := uc.links.DomainLink(ctx, userID, sphereID, "workspace")
	if err != nil {
		return SphereDomainLinkDTO{}, false, err
	}
	if !found {
		return SphereDomainLinkDTO{}, false, nil
	}
	return SphereDomainLinkDTO{LinkType: "workspace", RefID: refID, RefName: refName}, true, nil
}

// ListCareerContacts — имена контактов для бейджей в UI Карьеры.
type ListCareerContacts struct {
	lister CareerContactLister
}

func NewListCareerContacts(lister CareerContactLister) *ListCareerContacts {
	return &ListCareerContacts{lister: lister}
}

func (uc *ListCareerContacts) Execute(ctx context.Context, userID ids.UserID) ([]string, error) {
	if userID.IsZero() {
		return nil, fmt.Errorf("user id is required")
	}
	return uc.lister.ListContactNames(ctx, userID)
}
