/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	"context"
	"fmt"
	"time"

	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbtr "bscp.io/pkg/protocol/core/template-revision"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/search"
	"bscp.io/pkg/types"
)

// CreateTemplateRevision create template revision.
func (s *Service) CreateTemplateRevision(ctx context.Context, req *pbds.CreateTemplateRevisionReq) (*pbds.CreateResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if _, err := s.dao.TemplateRevision().GetByUniqueKey(kt, req.Attachment.BizId, req.Attachment.TemplateId,
		req.Spec.RevisionName); err == nil {
		return nil, fmt.Errorf("template revision's same revision name %s already exists", req.Spec.RevisionName)
	}

	template, err := s.dao.Template().GetByID(kt, req.Attachment.BizId, req.Attachment.TemplateId)
	if err != nil {
		logs.Errorf("get template by id failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	tx := s.dao.GenQuery().Begin()

	// 1. create template revision
	spec := req.Spec.TemplateRevisionSpec()
	spec.RevisionName = generateRevisionName()
	// keep the revision's name and path same with template
	spec.Name = template.Spec.Name
	spec.Path = template.Spec.Path
	templateRevision := &table.TemplateRevision{
		Spec:       spec,
		Attachment: req.Attachment.TemplateRevisionAttachment(),
		Revision: &table.CreatedRevision{
			Creator: kt.User,
		},
	}
	id, err := s.dao.TemplateRevision().CreateWithTx(kt, tx, templateRevision)
	if err != nil {
		logs.Errorf("create template revision failed, err: %v, rid: %s", err, kt.Rid)
		tx.Rollback()
		return nil, err
	}

	// 2. update the latest template revision for app template bindings if necessary
	atbs, err := s.dao.TemplateBindingRelation().
		ListLatestTemplateBoundUnnamedAppDetails(kt, req.Attachment.BizId, req.Attachment.TemplateId)
	if err != nil {
		logs.Errorf("create template revision failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	if len(atbs) > 0 {
		for _, atb := range atbs {
			updateBindingsWithLatest(atb.Spec.Bindings, req.Attachment.TemplateId, id)
			// no need to validate the template revision which just created
			// the query operation in genFinalATB method are not in the same transaction with the update operation
			// query it will cause error: `template release id is not exist`
			if err := s.genFinalATBWithTx(kt, tx, atb, []uint32{req.Attachment.TemplateId}); err != nil {
				logs.Errorf("create template revision failed, err: %v, rid: %s", err, kt.Rid)
				tx.Rollback()
				return nil, err
			}
			if err := s.dao.AppTemplateBinding().UpdateWithTx(kt, tx, atb); err != nil {
				logs.Errorf("create template revision failed, err: %v, rid: %s", err, kt.Rid)
				tx.Rollback()
				return nil, err
			}
		}
	}

	tx.Commit()

	resp := &pbds.CreateResp{Id: id}
	return resp, nil
}

// updateBindingsWithLatest get the final template bindings after update with the latest template revision
func updateBindingsWithLatest(bindings []*table.TemplateBinding, tmplID, latestRevisionID uint32) {
	for _, b := range bindings {
		for _, r := range b.TemplateRevisions {
			if r.TemplateID == tmplID {
				// update the latest template revision
				r.TemplateRevisionID = latestRevisionID
			}
		}
	}
}

// ListTemplateRevisions list template revision.
func (s *Service) ListTemplateRevisions(ctx context.Context, req *pbds.ListTemplateRevisionsReq) (*pbds.ListTemplateRevisionsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	opt := &types.BasePage{Start: req.Start, Limit: uint(req.Limit), All: req.All}
	if err := opt.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	searcher, err := search.NewSearcher(req.SearchFields, req.SearchValue, search.TemplateRevision)
	if err != nil {
		return nil, err
	}

	details, count, err := s.dao.TemplateRevision().List(kt, req.BizId, req.TemplateId, searcher, opt)

	if err != nil {
		logs.Errorf("list template revisions failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListTemplateRevisionsResp{
		Count:   uint32(count),
		Details: pbtr.PbTemplateRevisions(details),
	}
	return resp, nil
}

// DeleteTemplateRevision delete template revision.
func (s *Service) DeleteTemplateRevision(ctx context.Context, req *pbds.DeleteTemplateRevisionReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	templateRevision := &table.TemplateRevision{
		ID:         req.Id,
		Attachment: req.Attachment.TemplateRevisionAttachment(),
	}
	if err := s.dao.TemplateRevision().Delete(kt, templateRevision); err != nil {
		logs.Errorf("delete template revision failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// ListTemplateRevisionsByIDs list template revision by ids.
func (s *Service) ListTemplateRevisionsByIDs(ctx context.Context, req *pbds.ListTemplateRevisionsByIDsReq) (*pbds.
	ListTemplateRevisionsByIDsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTemplateRevisionsExist(kt, req.Ids); err != nil {
		return nil, err
	}

	details, err := s.dao.TemplateRevision().ListByIDs(kt, req.Ids)
	if err != nil {
		logs.Errorf("list template revisions failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListTemplateRevisionsByIDsResp{
		Details: pbtr.PbTemplateRevisions(details),
	}
	return resp, nil
}

// ListTemplateRevisionNamesByTemplateIDs list template revision by ids.
func (s *Service) ListTemplateRevisionNamesByTemplateIDs(ctx context.Context,
	req *pbds.ListTemplateRevisionNamesByTemplateIDsReq) (
	*pbds.
		ListTemplateRevisionNamesByTemplateIDsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTemplatesExist(kt, req.TemplateIds); err != nil {
		return nil, err
	}

	tmplRevisions, err := s.dao.TemplateRevision().ListByTemplateIDs(kt, req.BizId, req.TemplateIds)
	if err != nil {
		logs.Errorf("list template revision names by template ids failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	if len(tmplRevisions) == 0 {
		return &pbds.ListTemplateRevisionNamesByTemplateIDsResp{}, nil
	}

	// get the map of template id => the latest template revision id
	latestRevisionMap := make(map[uint32]uint32)
	// get the map of template id => template revision detail
	tmplRevisionMap := make(map[uint32]*pbtr.TemplateRevisionNamesDetail)
	for _, t := range tmplRevisions {
		if _, ok := latestRevisionMap[t.Attachment.TemplateID]; !ok {
			latestRevisionMap[t.Attachment.TemplateID] = t.ID
		} else {
			if t.ID > latestRevisionMap[t.Attachment.TemplateID] {
				latestRevisionMap[t.Attachment.TemplateID] = t.ID
			}
		}

		if _, ok := tmplRevisionMap[t.Attachment.TemplateID]; !ok {
			tmplRevisionMap[t.Attachment.TemplateID] = &pbtr.TemplateRevisionNamesDetail{}
		}
		tmplRevisionMap[t.Attachment.TemplateID].TemplateRevisions = append(
			tmplRevisionMap[t.Attachment.TemplateID].TemplateRevisions,
			&pbtr.TemplateRevisionNamesDetailRevisionNames{
				TemplateRevisionId:   t.ID,
				TemplateRevisionName: t.Spec.RevisionName,
			})
	}
	tmplIDs := make([]uint32, 0, len(tmplRevisionMap))
	for tmplID := range tmplRevisionMap {
		tmplIDs = append(tmplIDs, tmplID)
	}

	tmpls, err := s.dao.Template().ListByIDs(kt, tmplIDs)
	if err != nil {
		logs.Errorf("list template sets of biz failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	details := make([]*pbtr.TemplateRevisionNamesDetail, 0)
	for _, t := range tmpls {
		details = append(details, &pbtr.TemplateRevisionNamesDetail{
			TemplateId:               t.ID,
			TemplateName:             t.Spec.Name,
			LatestTemplateRevisionId: latestRevisionMap[t.ID],
			TemplateRevisions:        tmplRevisionMap[t.ID].TemplateRevisions,
		})
	}

	resp := &pbds.ListTemplateRevisionNamesByTemplateIDsResp{
		Details: details,
	}
	return resp, nil
}

func generateRevisionName() string {
	// Format the current time as YYYYMMDDHHMMSS
	timestamp := time.Now().Format("20060102150405")
	return fmt.Sprintf("v%s", timestamp)
}
