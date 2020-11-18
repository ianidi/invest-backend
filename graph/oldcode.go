package graph

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
// func (r *mutationResolver) OnboardingSignContract(ctx context.Context, input model.Signature) (*model.Onboarding, error) {
// 	// db := db.GetDB()

// 	var err error
// 	var onboarding *model.Onboarding

// 	// deepcopier.Copy(identity.Onboarding).To(onboarding)

// 	sender, err := QueryMember(ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	onboarding, err = QueryOnboardingRecord(sender.MemberID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	//Check that member didn't complete onboarding
// 	if onboarding.Complete {
// 		return onboarding, nil
// 	}

// 	encodedImage := strings.Split(input.Image, ",")

// 	if encodedImage[0] != "data:image/svg+xml;base64" {
// 		return nil, errors.New("INVALID_IMAGE")
// 	}

// 	decodedImage, err := base64.StdEncoding.DecodeString(encodedImage[1])
// 	if err != nil {
// 		return nil, err
// 	}

// 	if !utils.IsSVG(decodedImage) {
// 		return nil, errors.New("INVALID_IMAGE")
// 	}

// 	filename := cast.ToString(sender.MemberID) + "_" + uuid.New().String() + ".svg"

// 	_, err = s3.Upload(filename, decodedImage)
// 	if err != nil {
// 		return nil, errors.New("UPLOAD_ERROR")
// 	}

// 	return onboarding, nil
// }

// var category []models.Category
// var res []*model.InvestResponse

// var investCategory []*model.InvestCategory

// if err := db.Select(&category, "SELECT * FROM Category"); err != nil {
// 	if err != sql.ErrNoRows {
// 		return nil, err
// 	}
// }

// for _, categoryRow := range category {

// 	var invest []models.Invest

// 	if err := db.Select(&invest, "SELECT * FROM Invest WHERE CategoryID=$1", categoryRow.CategoryID); err != nil {
// 		if err != sql.ErrNoRows {
// 			return nil, err
// 		}
// 	}

// 	r := &model.InvestCategory{
// 		CategoryID: cast.ToInt(categoryRow.CategoryID),
// 		Title:      categoryRow.Title,
// 	}

// 	investCategory = append(investCategory, r)

// 	resRow := &model.InvestResponse{
// 		Category: {
// 			CategoryID: categoryRow.CategoryID,
// 			Title:      categoryRow.Title,
// 		},
// 	}

// }

// for _, investRow := range invest {

// 	resRow := &model.Invest{
// 		Title: investRow.Title,
// 	}

// 	res = append(res, resRow)
// }

// return res, nil
