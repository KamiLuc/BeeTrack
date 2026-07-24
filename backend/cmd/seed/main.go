// Command seed populates the database with a large set of realistic test
// data (apiaries, hives, inspections, treatments, harvests, and marketplace
// listings) owned by a single user, identified by email/password. Run it
// repeatedly to keep piling on more data — it never deletes anything.
//
// Usage (from backend/, with `docker compose up` already running):
//
//	go run ./cmd/seed -email=test@example.com -password=password123
//
// Drop 1-3 photos (jpg/png/webp) into cmd/seed/images/ beforehand if you
// want the seeded listings to have real images instead of none.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/beetrack/backend/internal/config"
	"github.com/beetrack/backend/internal/database"
	"github.com/beetrack/backend/internal/model"
	"github.com/beetrack/backend/internal/repository"
	"github.com/beetrack/backend/migrations"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

const maxImagesPerListing = 3

func main() {
	email := flag.String("email", "", "email of the account to seed data for (created if it doesn't exist)")
	password := flag.String("password", "", "password to (re)set on that account")
	apiURL := flag.String("api-url", "", "base URL of the running API, for image uploads (defaults to API_URL / http://localhost:8080)")
	imagesDir := flag.String("images-dir", "cmd/seed/images", "directory of jpg/png/webp files to attach to seeded listings")
	flag.Parse()

	if *email == "" || *password == "" {
		fmt.Fprintln(os.Stderr, "usage: go run ./cmd/seed -email=you@example.com -password=yourpassword")
		os.Exit(1)
	}

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if *apiURL == "" {
		*apiURL = cfg.APIURL
	}

	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	if err := database.Migrate(db, migrations.FS); err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	userRepo := repository.NewUserRepository(db)
	apiaryRepo := repository.NewApiaryRepository(db)
	hiveRepo := repository.NewHiveRepository(db)
	inspectionRepo := repository.NewInspectionRepository(db)
	treatmentRepo := repository.NewTreatmentRepository(db)
	feedingRepo := repository.NewFeedingRepository(db)
	harvestRepo := repository.NewHarvestRepository(db)
	listingRepo := repository.NewListingRepository(db)
	honeyBatchRepo := repository.NewHoneyBatchRepository(db)

	user, err := ensureUser(ctx, userRepo, *email, *password)
	if err != nil {
		log.Fatalf("ensure user: %v", err)
	}
	log.Printf("seeding data for user %q (id=%d)", user.Email, user.ID)

	apiaries := seedApiaries(ctx, apiaryRepo, user.ID)

	var allHives []*model.Hive
	allHives = append(allHives, seedHives(ctx, hiveRepo, apiaries[0], 6, 0)...)
	forestHives := seedHives(ctx, hiveRepo, apiaries[1], 5, 0)
	allHives = append(allHives, forestHives...)
	specialHives := seedSpecialHives(ctx, hiveRepo, apiaries[1], len(forestHives))
	allHives = append(allHives, specialHives...)
	if err := hiveRepo.CreateDisease(ctx, &model.HiveDisease{HiveID: specialHives[0].ID, Disease: "varroa"}); err != nil {
		log.Fatalf("create hive disease: %v", err)
	}

	inspectionCount, treatmentCount, feedingCount, harvestCount := 0, 0, 0, 0
	for i, h := range allHives {
		inspectionCount += seedInspections(ctx, inspectionRepo, h, user.ID, i, h.Name == "Sick")
		if i%2 == 0 {
			treatmentCount += seedTreatments(ctx, treatmentRepo, h, user.ID, i)
		}
		if h.Active {
			feedingCount += seedFeedings(ctx, feedingRepo, h, user.ID, i)
		}
		if i%2 == 1 || h.ReadyForHarvest {
			harvestCount += seedHarvests(ctx, harvestRepo, h, user.ID, i)
		}
	}
	log.Printf("created %d hives, %d inspections, %d treatments, %d feedings, %d harvests", len(allHives), inspectionCount, treatmentCount, feedingCount, harvestCount)

	// Pasieka Górska gets a handful of hives with a deliberately deep history
	// (many more records than the generic hives above) for exercising the
	// dashboard report screen and long history lists in the app.
	gorskaHives := seedGorskaHives(ctx, hiveRepo, apiaries[2])
	richInspectionCount, richTreatmentCount, richFeedingCount, richHarvestCount := 0, 0, 0, 0
	for _, h := range gorskaHives {
		richInspectionCount += seedManyInspections(ctx, inspectionRepo, h, user.ID, 30, 7)
		richFeedingCount += seedManyFeedings(ctx, feedingRepo, h, user.ID, 20, 10)
		richHarvestCount += seedManyHarvests(ctx, harvestRepo, h, user.ID, 5, 45)
		richTreatmentCount += seedManyTreatments(ctx, treatmentRepo, h, user.ID, 5, 35)
	}
	log.Printf("created %d Górska hives with rich history: %d inspections, %d treatments, %d feedings, %d harvests",
		len(gorskaHives), richInspectionCount, richTreatmentCount, richFeedingCount, richHarvestCount)

	listings, approveCount := seedListings(ctx, listingRepo, user.ID, apiaries, *email)

	seedHoneyBatches(ctx, honeyBatchRepo, user.ID)

	images := listImageFiles(*imagesDir)
	if len(images) == 0 {
		log.Printf("no images found in %s — listings created without photos, none approved", *imagesDir)
		return
	}

	token, err := login(*apiURL, *email, *password)
	if err != nil {
		log.Fatalf("login for image upload: %v", err)
	}
	uploaded := 0
	for i, l := range listings {
		n, err := uploadImages(*apiURL, token, l.ID, images, i)
		if err != nil {
			log.Printf("upload images for listing %d: %v", l.ID, err)
			continue
		}
		uploaded += n
	}
	log.Printf("uploaded %d listing photos", uploaded)

	// Approve only after photos are attached — uploading a photo resets an
	// already-approved listing back to pending, so approving first would be
	// immediately undone by the upload loop above.
	approved := 0
	for _, l := range listings[:approveCount] {
		if err := listingRepo.Approve(ctx, l.ID, user.ID); err != nil {
			log.Fatalf("approve listing %q: %v", l.Title, err)
		}
		approved++
	}
	log.Printf("approved %d of %d listings", approved, len(listings))
}

func ensureUser(ctx context.Context, repo *repository.UserRepository, email, password string) (*model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	existing, err := repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		if err := repo.UpdatePassword(ctx, existing.ID, string(hash)); err != nil {
			return nil, err
		}
		if !existing.Verified {
			if err := repo.SetVerified(ctx, existing.ID); err != nil {
				return nil, err
			}
		}
		return existing, nil
	}

	user := &model.User{
		Email:        email,
		Name:         "Testowy Pszczelarz",
		PasswordHash: string(hash),
		Verified:     true,
	}
	if err := repo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func seedApiaries(ctx context.Context, repo *repository.ApiaryRepository, ownerID int64) []*model.Apiary {
	krakowLat, krakowLng := 50.0647, 19.9450
	lesnaLat, lesnaLng := 50.2649, 19.0238
	zakopaneLat, zakopaneLng := 49.2992, 19.9496
	specs := []*model.Apiary{
		{OwnerUserID: ownerID, Name: "Pasieka Słoneczna", Lat: &krakowLat, Lng: &krakowLng, GridRows: 4, GridCols: 5},
		{OwnerUserID: ownerID, Name: "Pasieka Leśna", Lat: &lesnaLat, Lng: &lesnaLng, GridRows: 2, GridCols: 10},
		{OwnerUserID: ownerID, Name: "Pasieka Górska", Lat: &zakopaneLat, Lng: &zakopaneLng, GridRows: 3, GridCols: 3},
	}
	for _, a := range specs {
		if err := repo.Create(ctx, a, "owner"); err != nil {
			log.Fatalf("create apiary %q: %v", a.Name, err)
		}
	}
	return specs
}

var hiveTypes = []string{"Wielkopolski", "Dadant", "Warszawski zwykły", "Langstroth"}

// seedHives creates count generically-named hives in apiary, filling grid positions
// row-major starting at startIdx (so callers can reserve trailing positions for
// seedSpecialHives without colliding).
func seedHives(ctx context.Context, repo *repository.HiveRepository, apiary *model.Apiary, count, startIdx int) []*model.Hive {
	var hives []*model.Hive
	for i := 0; i < count; i++ {
		idx := startIdx + i
		h := &model.Hive{
			ApiaryID:        apiary.ID,
			Name:            fmt.Sprintf("Ul %d", idx+1),
			Type:            hiveTypes[idx%len(hiveTypes)],
			Active:          idx%5 != 2,
			Queenless:       idx%5 == 2,
			ReadyForHarvest: idx%5 == 4,
			NeedsFood:       idx%5 == 3,
			GridRow:         idx / apiary.GridCols,
			GridCol:         idx % apiary.GridCols,
		}
		if err := repo.Create(ctx, h); err != nil {
			log.Fatalf("create hive %q: %v", h.Name, err)
		}
		hives = append(hives, h)
	}
	return hives
}

// seedSpecialHives creates three hives with fixed names for exercising specific
// states in the app: one sick (flagged separately via a HiveDisease by the
// caller), one ready for harvest, and one queenless.
func seedSpecialHives(ctx context.Context, repo *repository.HiveRepository, apiary *model.Apiary, startIdx int) []*model.Hive {
	specs := []struct {
		name            string
		active          bool
		queenless       bool
		readyForHarvest bool
		needsFood       bool
	}{
		{"Sick", true, false, false, false},
		{"Ready", true, false, true, false},
		{"Queenless", true, true, false, false},
		{"Needs Food", true, false, false, true},
	}
	var hives []*model.Hive
	for i, s := range specs {
		idx := startIdx + i
		h := &model.Hive{
			ApiaryID:        apiary.ID,
			Name:            s.name,
			Type:            hiveTypes[idx%len(hiveTypes)],
			Active:          s.active,
			Queenless:       s.queenless,
			ReadyForHarvest: s.readyForHarvest,
			NeedsFood:       s.needsFood,
			GridRow:         idx / apiary.GridCols,
			GridCol:         idx % apiary.GridCols,
		}
		if err := repo.Create(ctx, h); err != nil {
			log.Fatalf("create hive %q: %v", h.Name, err)
		}
		hives = append(hives, h)
	}
	return hives
}

// seedGorskaHives creates three plain, always-active hives in apiary for
// exercising a hive with a long, dense history (see seedMany* below).
func seedGorskaHives(ctx context.Context, repo *repository.HiveRepository, apiary *model.Apiary) []*model.Hive {
	names := []string{"Ul Górski 1", "Ul Górski 2", "Ul Górski 3"}
	var hives []*model.Hive
	for i, name := range names {
		h := &model.Hive{
			ApiaryID: apiary.ID,
			Name:     name,
			Type:     hiveTypes[i%len(hiveTypes)],
			Active:   true,
			GridRow:  i / apiary.GridCols,
			GridCol:  i % apiary.GridCols,
		}
		if err := repo.Create(ctx, h); err != nil {
			log.Fatalf("create hive %q: %v", h.Name, err)
		}
		hives = append(hives, h)
	}
	return hives
}

var queenStatuses = []string{"seen", "not_seen"}
var broodPatterns = []string{"excellent", "good", "poor", "none"}
var aggressivenessLevels = []string{"calm", "mild", "aggressive", "very_aggressive"}
var inspectionNotes = []string{
	"Rodzina silna, matka czerwi równomiernie.",
	"Widoczne oznaki rójki, dołożono ramkę z zasuszem.",
	"Zapasy pyłku w normie, brak niepokojących objawów.",
	"Nieco osłabiona rodzina po zimie, obserwować.",
	"Matka bardzo płodna, plastry pełne czerwiu.",
}
var diseases = []string{"varroa", "nosema", "chalkbrood"}

// seedInspections creates three inspections for hive, spread over the last two
// months. When forceDisease is true (the dedicated "Sick" hive), the most
// recent inspection always gets a disease record; otherwise diseases show up
// occasionally for variety.
func seedInspections(ctx context.Context, repo *repository.InspectionRepository, hive *model.Hive, userID int64, hiveIdx int, forceDisease bool) int {
	count := 0
	for j, daysAgo := range []int{60, 30, 7} {
		idx := hiveIdx + j
		framesBrood := 3 + idx%6
		framesFeed := 1 + idx%4
		framesPollen := 1 + idx%3
		insp := &model.Inspection{
			HiveID:         hive.ID,
			InspectedBy:    userID,
			InspectedAt:    time.Now().AddDate(0, 0, -daysAgo),
			QueenStatus:    queenStatuses[idx%len(queenStatuses)],
			BroodPattern:   broodPatterns[idx%len(broodPatterns)],
			FramesBrood:    &framesBrood,
			FramesFeed:     &framesFeed,
			FramesPollen:   &framesPollen,
			Aggressiveness: aggressivenessLevels[idx%len(aggressivenessLevels)],
			Notes:          inspectionNotes[idx%len(inspectionNotes)],
		}
		if err := repo.Create(ctx, insp); err != nil {
			log.Fatalf("create inspection for hive %d: %v", hive.ID, err)
		}
		count++

		isLast := j == 2
		if (forceDisease && isLast) || (!forceDisease && hiveIdx%3 == 0 && isLast) {
			disease := diseases[hiveIdx%len(diseases)]
			if forceDisease {
				disease = "varroa"
			}
			if err := repo.CreateDisease(ctx, &model.InspectionDisease{
				InspectionID: insp.ID,
				Disease:      disease,
				Notes:        "Wykryto podczas przeglądu, wdrożono leczenie.",
			}); err != nil {
				log.Fatalf("create inspection disease: %v", err)
			}
		}
	}
	return count
}

// seedManyInspections creates count inspections for hive, one every
// intervalDays going back from today — for hives that need a long, dense
// history rather than the standard three-inspection sample.
func seedManyInspections(ctx context.Context, repo *repository.InspectionRepository, hive *model.Hive, userID int64, count, intervalDays int) int {
	for j := 0; j < count; j++ {
		framesBrood := 3 + j%6
		framesFeed := 1 + j%4
		framesPollen := 1 + j%3
		insp := &model.Inspection{
			HiveID:         hive.ID,
			InspectedBy:    userID,
			InspectedAt:    time.Now().AddDate(0, 0, -j*intervalDays),
			QueenStatus:    queenStatuses[j%len(queenStatuses)],
			BroodPattern:   broodPatterns[j%len(broodPatterns)],
			FramesBrood:    &framesBrood,
			FramesFeed:     &framesFeed,
			FramesPollen:   &framesPollen,
			Aggressiveness: aggressivenessLevels[j%len(aggressivenessLevels)],
			Notes:          inspectionNotes[j%len(inspectionNotes)],
		}
		if err := repo.Create(ctx, insp); err != nil {
			log.Fatalf("create inspection for hive %d: %v", hive.ID, err)
		}
	}
	return count
}

var medicines = []string{"Apiwarol", "Kwas szczawiowy", "Bayvarol", "MAQS"}
var doses = []string{"1 pasek/ul", "5 ml", "2 paski/ul", "1 saszetka"}

func seedTreatments(ctx context.Context, repo *repository.TreatmentRepository, hive *model.Hive, userID int64, hiveIdx int) int {
	count := 0
	for j, daysAgo := range []int{40, 15} {
		idx := hiveIdx + j
		t := &model.Treatment{
			HiveID:       hive.ID,
			TreatedBy:    userID,
			TreatedAt:    time.Now().AddDate(0, 0, -daysAgo),
			MedicineName: medicines[idx%len(medicines)],
			Dose:         doses[idx%len(doses)],
			Notes:        "Zabieg przeciw warrozie, kontynuować zgodnie z planem.",
		}
		if err := repo.Create(ctx, t); err != nil {
			log.Fatalf("create treatment for hive %d: %v", hive.ID, err)
		}
		count++
	}
	return count
}

// seedManyTreatments is the seedTreatments equivalent for hives that need a
// long, dense history — see seedManyInspections.
func seedManyTreatments(ctx context.Context, repo *repository.TreatmentRepository, hive *model.Hive, userID int64, count, intervalDays int) int {
	for j := 0; j < count; j++ {
		t := &model.Treatment{
			HiveID:       hive.ID,
			TreatedBy:    userID,
			TreatedAt:    time.Now().AddDate(0, 0, -j*intervalDays),
			MedicineName: medicines[j%len(medicines)],
			Dose:         doses[j%len(doses)],
			Notes:        "Zabieg przeciw warrozie, kontynuować zgodnie z planem.",
		}
		if err := repo.Create(ctx, t); err != nil {
			log.Fatalf("create treatment for hive %d: %v", hive.ID, err)
		}
	}
	return count
}

var feedTypes = []string{"Syrop cukrowy 1:1", "Ciasto cukrowe", "Pokarm inwertowany", "Syrop cukrowy 3:2"}
var feedAmounts = []string{"2 l", "1 kg", "1.5 l", "0.5 kg"}

func seedFeedings(ctx context.Context, repo *repository.FeedingRepository, hive *model.Hive, userID int64, hiveIdx int) int {
	count := 0
	for j, daysAgo := range []int{25, 10} {
		idx := hiveIdx + j
		f := &model.Feeding{
			HiveID:   hive.ID,
			FedBy:    userID,
			FedAt:    time.Now().AddDate(0, 0, -daysAgo),
			FeedType: feedTypes[idx%len(feedTypes)],
			Amount:   feedAmounts[idx%len(feedAmounts)],
			Notes:    "Podkarmianie w ramach przygotowań rodziny.",
		}
		if err := repo.Create(ctx, f); err != nil {
			log.Fatalf("create feeding for hive %d: %v", hive.ID, err)
		}
		count++
	}
	return count
}

// seedManyFeedings is the seedFeedings equivalent for hives that need a
// long, dense history — see seedManyInspections.
func seedManyFeedings(ctx context.Context, repo *repository.FeedingRepository, hive *model.Hive, userID int64, count, intervalDays int) int {
	for j := 0; j < count; j++ {
		f := &model.Feeding{
			HiveID:   hive.ID,
			FedBy:    userID,
			FedAt:    time.Now().AddDate(0, 0, -j*intervalDays),
			FeedType: feedTypes[j%len(feedTypes)],
			Amount:   feedAmounts[j%len(feedAmounts)],
			Notes:    "Podkarmianie w ramach przygotowań rodziny.",
		}
		if err := repo.Create(ctx, f); err != nil {
			log.Fatalf("create feeding for hive %d: %v", hive.ID, err)
		}
	}
	return count
}

func seedHarvests(ctx context.Context, repo *repository.HarvestRepository, hive *model.Hive, userID int64, hiveIdx int) int {
	frames := 6 + hiveIdx%5
	kilograms := 8.5 + float64(hiveIdx%7)*1.5
	h := &model.Harvest{
		HiveID:      hive.ID,
		HarvestedBy: userID,
		HarvestedAt: time.Now().AddDate(0, 0, -20),
		Frames:      frames,
		HalfFrames:  hiveIdx % 3,
		Kilograms:   kilograms,
		Notes:       "Miód wielokwiatowy, dobra jakość, niska wilgotność.",
	}
	if err := repo.Create(ctx, h); err != nil {
		log.Fatalf("create harvest for hive %d: %v", hive.ID, err)
	}
	return 1
}

// seedManyHarvests is the seedHarvests equivalent for hives that need a
// long, dense history — see seedManyInspections.
func seedManyHarvests(ctx context.Context, repo *repository.HarvestRepository, hive *model.Hive, userID int64, count, intervalDays int) int {
	for j := 0; j < count; j++ {
		frames := 6 + j%5
		kilograms := 8.5 + float64(j%7)*1.5
		h := &model.Harvest{
			HiveID:      hive.ID,
			HarvestedBy: userID,
			HarvestedAt: time.Now().AddDate(0, 0, -j*intervalDays),
			Frames:      frames,
			HalfFrames:  j % 3,
			Kilograms:   kilograms,
			Notes:       "Miód wielokwiatowy, dobra jakość, niska wilgotność.",
		}
		if err := repo.Create(ctx, h); err != nil {
			log.Fatalf("create harvest for hive %d: %v", hive.ID, err)
		}
	}
	return count
}

// cityCoords maps each city name used in the seeded listings' Address field to
// its real-world coordinates, so the distance filter has something meaningful
// to filter/sort by.
var cityCoords = map[string][2]float64{
	"Kraków":    {50.0647, 19.9450},
	"Wieliczka": {49.9880, 20.0561},
	"Tarnów":    {50.0121, 20.9858},
	"Bochnia":   {49.9702, 20.4310},
	"Zakopane":  {49.2992, 19.9496},
	"Lublin":    {51.2465, 22.5684},
	"Nowy Sącz": {49.6221, 20.6906},
}

var quantityCountRe = regexp.MustCompile(`^(\d+)(.*)$`)

// jitterQuantity randomly varies the leading count in a quantity string like
// "10 słoików 0.9kg" by up to ±40% (minimum 1), so repeated seed runs don't
// produce visually identical listings.
func jitterQuantity(q string) string {
	m := quantityCountRe.FindStringSubmatch(q)
	if m == nil {
		return q
	}
	n, err := strconv.Atoi(m[1])
	if err != nil {
		return q
	}
	delta := int(float64(n) * 0.4)
	if delta < 1 {
		delta = 1
	}
	jittered := n - delta + rand.Intn(2*delta+1)
	if jittered < 1 {
		jittered = 1
	}
	return fmt.Sprintf("%d%s", jittered, m[2])
}

// jitterPrice randomly varies base by ±15%, rounded to the nearest 0.50.
func jitterPrice(base float64) float64 {
	factor := 0.85 + rand.Float64()*0.3
	return math.Round(base*factor*2) / 2
}

// jitterCoord nudges (lat, lng) by up to ~3km in a random direction, so
// seeded listings in the same city aren't all pinned to the exact same point.
func jitterCoord(lat, lng float64) (float64, float64) {
	const maxDeltaDeg = 0.03
	jitter := func(v float64) float64 { return v + (rand.Float64()*2-1)*maxDeltaDeg }
	return jitter(lat), jitter(lng)
}

// seedListings creates all listings (both to-be-approved and left-pending)
// and returns them with the to-be-approved ones first — approveCount is how
// many of those leading entries to approve, once the caller has uploaded
// their photos.
func seedListings(ctx context.Context, repo *repository.ListingRepository, userID int64, apiaries []*model.Apiary, email string) (listings []*model.Listing, approveCount int) {
	price := func(v float64) *float64 { return &v }
	specs := []*model.Listing{
		{
			Title: "Miód wielokwiatowy 2026", Description: "Miód wielokwiatowy zebrany latem, nieprzegrzewany, tłoczony na zimno.",
			Category: "HONEY", Price: price(35), Quantity: "10 słoików 0.9kg", Address: "Kraków",
			ApiaryID: &apiaries[0].ID, ContactPhone: "+48 600 100 200", ContactEmail: email,
		},
		{
			Title: "Miód rzepakowy, świeży odwirunek", Description: "Kremowy miód rzepakowy, zbiór z tego sezonu.",
			Category: "HONEY", Price: price(30), Quantity: "5 słoików 0.9kg", Address: "Wieliczka",
			ContactPhone: "+48 600 100 200", ContactEmail: email,
		},
		{
			Title: "Ul wielkopolski, stan bardzo dobry", Description: "Używany ul wielkopolski, docieplony, gotowy do użytku.",
			Category: "BEEHIVES", Price: price(450), Quantity: "1 szt.", Address: "Kraków",
			ContactPhone: "+48 600 100 200", ContactEmail: email,
		},
		{
			Title: "Ule Dadant – zestaw 3 szt.", Description: "Komplet trzech uli Dadant, po sezonie, w dobrym stanie.",
			Category: "BEEHIVES", Price: price(1200), Quantity: "3 szt.", Address: "Tarnów",
			ContactPhone: "+48 600 100 200", ContactEmail: email,
		},
		{
			Title: "Rodzina pszczela na ramce wielkopolskiej", Description: "Silna rodzina z młodą matką z tego roku.",
			Category: "BEE_COLONIES", Price: price(550), Quantity: "1 rodzina", Address: "Kraków",
			ApiaryID: &apiaries[0].ID, ContactPhone: "+48 600 100 200", ContactEmail: email,
		},
		{
			Title: "Odkłady pszczele – silne rodziny", Description: "Odkłady na 5 ramkach, zdrowe i silne rodziny.",
			Category: "BEE_COLONIES", Price: price(480), Quantity: "2 rodziny", Address: "Bochnia",
			ContactPhone: "+48 600 100 200", ContactEmail: email,
		},
		{
			Title: "Miód spadziowy leśny", Description: "Ciemny miód spadziowy z lasów iglastych, bogaty w minerały.",
			Category: "HONEY", Price: price(45), Quantity: "8 słoików 0.9kg", Address: "Zakopane",
			ApiaryID: &apiaries[2].ID, ContactPhone: "+48 600 100 200", ContactEmail: email,
		},
		{
			Title: "Miód gryczany", Description: "Intensywny miód gryczany, polecany na przeziębienie.",
			Category: "HONEY", Price: price(32), Quantity: "6 słoików 0.9kg", Address: "Lublin",
			ContactPhone: "+48 600 100 200", ContactEmail: email,
		},
		{
			Title: "Ul Langstroth, ocieplony styropianem", Description: "Ul Langstroth z ociepleniem, po jednym sezonie użytkowania.",
			Category: "BEEHIVES", Price: price(380), Quantity: "1 szt.", Address: "Nowy Sącz",
			ContactPhone: "+48 600 100 200", ContactEmail: email,
		},
		{
			Title: "Ule warszawskie zwykłe – zestaw 2 szt.", Description: "Dwa ule warszawskie zwykłe, solidna konstrukcja.",
			Category: "BEEHIVES", Price: price(700), Quantity: "2 szt.", Address: "Kraków",
			ContactPhone: "+48 600 100 200", ContactEmail: email,
		},
		{
			Title: "Rodziny pszczele na zimowlę", Description: "Rodziny przygotowane do zimowli, dobrze zaopatrzone.",
			Category: "BEE_COLONIES", Price: price(600), Quantity: "3 rodziny", Address: "Tarnów",
			ContactPhone: "+48 600 100 200", ContactEmail: email,
		},
		{
			Title: "Pakiety pszczele 1.5kg z matką", Description: "Pakiety pszczele z młodą, unasienioną matką.",
			Category: "BEE_COLONIES", Price: price(420), Quantity: "4 pakiety", Address: "Kraków",
			ApiaryID: &apiaries[1].ID, ContactPhone: "+48 600 100 200", ContactEmail: email,
		},
	}
	// Left pending (not auto-approved below) so the admin review queue has
	// something to show right after seeding, without a human creating one by hand.
	pendingSpecs := []*model.Listing{
		{
			Title: "Świeży miód akacjowy", Description: "Jasny miód akacjowy, dopiero co odwirowany, czeka na zatwierdzenie.",
			Category: "HONEY", Price: price(38), Quantity: "6 słoików 0.9kg", Address: "Kraków",
			ContactPhone: "+48 600 100 200", ContactEmail: email,
		},
		{
			Title: "Matki pszczele reprodukcyjne", Description: "Matki pszczele czystej rasy, tegoroczny chów.",
			Category: "QUEEN_BEES", Price: price(150), Quantity: "3 szt.", Address: "Wieliczka",
			ContactPhone: "+48 600 100 200", ContactEmail: email,
		},
	}

	// Only created here, left pending — approving happens in main(), after
	// photos are uploaded through the real API. Uploading a photo resets an
	// already-approved listing back to pending (see
	// ListingImageService.resetToPendingIfReviewed), so approving before the
	// upload step would just have the upload immediately undo it.
	for _, l := range append(specs, pendingSpecs...) {
		l.UserID = userID
		if coords, ok := cityCoords[l.Address]; ok {
			l.Lat, l.Lng = jitterCoord(coords[0], coords[1])
		}
		if l.Price != nil {
			jittered := jitterPrice(*l.Price)
			l.Price = &jittered
		}
		l.Quantity = jitterQuantity(l.Quantity)
		if err := repo.Create(ctx, l); err != nil {
			log.Fatalf("create listing %q: %v", l.Title, err)
		}
	}

	// Split evenly regardless of the specs/pendingSpecs split above, so the
	// admin review queue and the approved feed both have a realistic amount
	// to show — half the seeded listings waiting for approval, half live.
	all := append(specs, pendingSpecs...)
	approveCount = len(all) / 2
	log.Printf("created %d listings (%d to be approved, %d left pending review)", len(all), approveCount, len(all)-approveCount)
	return all, approveCount
}

var honeyBatchSpecs = []struct {
	honeyType        string
	processingMethod model.ProcessingMethod
	amountGrams      int64
	daysAgo          int
}{
	{"Wielokwiatowy", model.ProcessingMethodRaw, 12000, 45},
	{"Rzepakowy", model.ProcessingMethodFiltered, 8000, 30},
	{"Spadziowy", model.ProcessingMethodRaw, 6000, 20},
	{"Gryczany", model.ProcessingMethodPasteurized, 9500, 15},
	{"Akacjowy", model.ProcessingMethodFiltered, 5000, 5},
}

// seedHoneyBatches creates a handful of honey batches for userID with no
// certification requested, leaving them available for the owner (or admin
// review flows) to request certification on manually.
func seedHoneyBatches(ctx context.Context, repo *repository.HoneyBatchRepository, userID int64) int {
	for _, s := range honeyBatchSpecs {
		token, err := uuid.NewRandom()
		if err != nil {
			log.Fatalf("generate verification token: %v", err)
		}
		batch := &model.HoneyBatch{
			UserID:            userID,
			VerificationToken: token.String(),
			GatheringDate:     time.Now().AddDate(0, 0, -s.daysAgo),
			AmountGrams:       s.amountGrams,
			ProcessingMethod:  string(s.processingMethod),
			HoneyType:         s.honeyType,
		}
		if err := repo.CreateWithCertificationRequest(ctx, batch, nil); err != nil {
			log.Fatalf("create honey batch %q: %v", s.honeyType, err)
		}
	}
	log.Printf("created %d honey batches", len(honeyBatchSpecs))
	return len(honeyBatchSpecs)
}

var imageExtensions = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}

func listImageFiles(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() || !imageExtensions[strings.ToLower(filepath.Ext(e.Name()))] {
			continue
		}
		files = append(files, filepath.Join(dir, e.Name()))
	}
	sort.Strings(files)
	return files
}

func login(apiURL, email, password string) (string, error) {
	body, _ := json.Marshal(map[string]string{"email": email, "password": password})
	resp, err := http.Post(apiURL+"/api/v1/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("login failed (%d): %s", resp.StatusCode, data)
	}
	var out struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.AccessToken, nil
}

// uploadImages attaches up to maxImagesPerListing photos to listingID, cycling through
// files starting at offset so different listings get different photos when there
// are more listings than source images.
func uploadImages(apiURL, token string, listingID int64, files []string, offset int) (int, error) {
	n := len(files)
	if n > maxImagesPerListing {
		n = maxImagesPerListing
	}
	uploaded := 0
	var lastErr error
	for i := 0; i < n; i++ {
		path := files[(offset+i)%len(files)]
		if err := uploadImage(apiURL, token, listingID, path); err != nil {
			lastErr = err
			continue
		}
		uploaded++
	}
	return uploaded, lastErr
}

func uploadImage(apiURL, token string, listingID int64, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	mimeType := http.DetectContentType(data)
	if mimeType != "image/jpeg" && mimeType != "image/png" && mimeType != "image/webp" {
		return fmt.Errorf("%s: unsupported image type %s", path, mimeType)
	}

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="image"; filename=%q`, filepath.Base(path)))
	header.Set("Content-Type", mimeType)
	part, err := w.CreatePart(header)
	if err != nil {
		return err
	}
	if _, err := part.Write(data); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/v1/listings/%d/images", apiURL, listingID)
	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload %s failed (%d): %s", path, resp.StatusCode, body)
	}
	return nil
}
