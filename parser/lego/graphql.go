package lego

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/nbr23/atomic-banquet/parser"
)

func legoFeedQuery(options *parser.Options) (*http.Response, error) {

	query := `query ContentPageQuery($slug: String!, $perPage: Int, $page: Int, $sort: SortInput, $filters: [Filter!], $searchSessionId: Int, $offset: Int) {
    contentPage(slug: $slug) {
        id
        analyticsGroup
        analyticsPageTitle
        metaTitle
        metaDescription
        metaOpenGraph {
            title
            description
            imageUrl
            __typename
        }
        noIndexTag
        url
        title
        displayTitleOnPage
        ...Breadcrumbs
        articleTags {
            id
            title
            __typename
        }
        updatedAt
        createdAt
        jsonSchemaProperties {
            type
            __typename
        }
        contentBody {
            ...ContentContainerData_WithProductListingSections
            ...ContentContainerData_WithGiftFinderResultsSection
            __typename
        }
        __typename
    }
}

fragment Product_ProductItem on Product {
    __typename
    id
    productCode
    name
    slug
    primaryImage(size: THUMBNAIL)
    baseImgUrl: primaryImage
    additionalImages {
        url
        tag
        __typename
    }
    listingImages: listingAssets(type: IMAGE, limit: 2) {
        ... on ListingProductAsset {
            id
            tag
            url
            thumbnailDimensions {
                height
                width
                __typename
            }
            __typename
        }
        __typename
    }
    overrideUrl
    ... on ReadOnlyProduct {
        readOnlyVariant {
            ...Variant_ReadOnlyProduct
            __typename
        }
        __typename
    }
    ... on SingleVariantProduct {
        variant {
            ...Variant_ListingProduct
            __typename
        }
        __typename
    }
    ... on MultiVariantProduct {
        priceRange {
            formattedPriceRange
            formattedListPriceRange
            __typename
        }
        variants {
            ...Variant_ListingProduct
            __typename
        }
        __typename
    }
}

fragment Variant_ListingProduct on ProductVariant {
    id
    sku
    salePercentage
    attributes {
        rating
        maxOrderQuantity
        availabilityStatus
        availabilityText
        vipAvailabilityStatus
        vipAvailabilityText
        canAddToBag
        canAddToWishlist
        vipCanAddToBag
        onSale
        isNew
        ageRange
        pieceCount
        safetyWarning {
            safetyWarningKey: key
            safetyWarningLabel: label
            showSafetyImage
            imageKey
            __typename
        }
        ...ProductAttributes_Flags
        __typename
    }
    ...ProductVariant_Pricing
    __typename
}

fragment ProductVariant_Pricing on ProductVariant {
    price {
        formattedAmount
        centAmount
        currencyCode
        formattedValue
        __typename
    }
    listPrice {
        formattedAmount
        centAmount
        __typename
    }
    attributes {
        onSale
        __typename
    }
    __typename
}

fragment ProductAttributes_Flags on ProductAttributes {
    featuredFlags {
        key
        label
        __typename
    }
    __typename
}

fragment Variant_ReadOnlyProduct on ReadOnlyVariant {
    id
    sku
    attributes {
        featuredFlags {
            key
            label
            __typename
        }
        ageRange
        pieceCount
        __typename
    }
    __typename
}

fragment Breadcrumbs on Content {
    breadcrumbs {
        __typename
        label
        url
        analyticsTitle
    }
    __typename
}

fragment ContentContainerData_WithProductListingSections on ContentContainer {
    ...BaseContentContainer
    section {
        ...BaseContentSection
        ... on ProductSection {
            removePadding
            filterName
            filterValue
            ... on DisruptorProductSection {
                ...DisruptorSection
                __typename
            }
            products(perPage: $perPage
            page: $page
            sort: $sort
            filters: $filters
            searchSession: $searchSessionId
            offset: $offset) {
                ...ProductListings
                __typename
            }
            __typename
        }
        __typename
    }
    __typename
}

fragment DisruptorSection on DisruptorProductSection {
    disruptor {
        ...DisruptorData
        __typename
    }
    __typename
}

fragment DisruptorData on Disruptor {
    __typename
    imageSrc {
        ...ImageAsset
        __typename
    }
    contrast
    background
    title
    description
    link
    openInNewTab
}

fragment ImageAsset on ImageAssetDetails {
    url
    width
    height
    maxPixelDensity
    format
    __typename
}

fragment ProductListings on ProductQueryResult {
    resultId
    count
    offset
    perPage
    total
    optimizelyExperiment {
        testId
        variantId
        __typename
    }
    fetchDisruptorsOnClient
    reservedSlots
    results {
        ...Product_ProductItem
        __typename
    }
    facets {
        ...Facet_FacetSidebar
        __typename
    }
    sortOptions {
        ...Sort_SortOptions
        __typename
    }
    disruptors {
        ...DisruptorStatic
        ...DisruptorPopupVideo
        ...DisruptorInteractive
        __typename
    }
    __typename
}

fragment Facet_FacetSidebar on Facet {
    name
    key
    id
    isDefaultActive
    labels {
        __typename
        displayMode
        name
        labelKey
        count
        ... on FacetValue {
            value
            __typename
        }
        ... on FacetRange {
            from
            to
            __typename
        }
    }
    __typename
}

fragment Sort_SortOptions on SortOptions {
    id
    key
    direction
    label
    analyticLabel
    __typename
}

fragment DisruptorStatic on DisruptorStatic {
    __typename
    contentType
    title
    background
    colorPalette {
        designToken
        contrastingTextColor
        __typename
    }
    isDouble
    description
    url
    callToActionText
    image {
        url
        width
        height
        maxPixelDensity
        format
        __typename
    }
    logos {
        image {
            url
            width
            height
            maxPixelDensity
            format
            __typename
        }
        position {
            vertical
            horizontal
            __typename
        }
        altText
        __typename
    }
    testMetadata {
        experimentKey
        variationKey
        __typename
    }
}

fragment DisruptorInteractive on DisruptorInteractive {
    __typename
    title
    description
    background
    colorPalette {
        designToken
        contrastingTextColor
        __typename
    }
    url
    callToActionText
    contentType
    videoMedia {
        url
        __typename
    }
    logos {
        image {
            url
            width
            height
            maxPixelDensity
            format
            __typename
        }
        position {
            vertical
            horizontal
            __typename
        }
        altText
        __typename
    }
    thumbnailImage {
        url
        width
        height
        maxPixelDensity
        format
        __typename
    }
    testMetadata {
        experimentKey
        variationKey
        __typename
    }
}

fragment DisruptorPopupVideo on DisruptorPopupVideo {
    __typename
    contentType
    title
    description
    background
    colorPalette {
        designToken
        contrastingTextColor
        __typename
    }
    image {
        url
        width
        height
        maxPixelDensity
        format
        __typename
    }
    popupVideo {
        subtitlesUrl
        url
        __typename
    }
    logos {
        image {
            url
            width
            height
            maxPixelDensity
            format
            __typename
        }
        position {
            vertical
            horizontal
            __typename
        }
        altText
        __typename
    }
    testMetadata {
        experimentKey
        variationKey
        __typename
    }
}

fragment BaseContentContainer on ContentContainer {
    id
    multivariate {
        experimentId
        entityId
        testingId
        inExperimentAudience
        __typename
    }
    targeting {
        fetchOnClient
        __typename
    }
    __typename
}

fragment BaseContentSection on ContentSection {
    __typename
    id
    layout {
        width
        colors {
            background
            __typename
        }
        __typename
    }
    ...CarouselContentSection
    ...CustomCarouselContentSection
    ...UserGeneratedContentData
    ...AccordionSectionData
    ...BreadcrumbSection
    ...CategoryListingSection
    ...ListingBannerSection
    ...CardContentSection
    ...CardCarouselSection
    ...CopyContent
    ...CopySectionData
    ...QuickLinksData
    ...ContentBlockMixedData
    ...HeroBannerData
    ...MotionBannerData
    ...MotionSidekickData
    ...InPageNavData
    ...GalleryData
    ...TableData
    ...CountdownBannerData
    ...RecommendationSectionData
    ...SidekickBannerData
    ...TextBlockData
    ...TextBlockSEOData
    ...CrowdTwistWidgetSection
    ...CrowdTwistToggleWidgetSection
    ...CrowdTwistCodeRedemptionBanner
    ...CodedSection
    ...GridSectionData
    ...StickyCTAData
    ...AudioSectionData
    ...MotionSidekick1x1Data
    ...ImageTransitionSliderData
    ...ImageXrayViewerData
    ...PollsSectionData
    ...ArtNavigationData
    ...MotionBanner16x9Data
    ...QuickLinksAdvancedData
    ...ArticleSectionData
    ...RelatedArticleSectionData
    ...FeatureExplorerSectionData
    ...IdeaGeneratorSectionData
    ...TabbedContentExplorerData
    ...CustomProductCarousel_UniqueFields
    ...CustomProductCarousel_ItemFields
    ...CardContentRTWData
    ...ExpandedCardContentData
    ...ArticleTextData
    ...ArticleImageSectionData
    ...ExpandedProductLeafData
    ...Dots3DSectionData
    ...NinetiethAnniversaryExperienceData
    ...ArticleGroupSectionData
    ...MotionBannerSectionData
    ...AdvancedProductHeroBannerSectionData
    ...PlayTypeDetectorSectionData
    ...SocialShareSectionData
    ...EcosystemJourneyStarterData
    ...ResultsBannerSectionData
    ...CharacterExplorerSectionData
    ...SubmissionFormBannerSectionData
    ...RebuildTheWorld2023Data
}

fragment CarouselContentSection on ContentSection {
    ... on ProductCarouselSection {
        ...ProductCarousel_UniqueFields
        productCarouselProducts: products(page: 1
        perPage: 16
        sort: { key: FEATURED, direction: DESC }) {
            ...Product_ProductItem
            __typename
        }
        __typename
    }
    __typename
}

fragment CustomCarouselContentSection on ContentSection {
    ... on CustomProductCarouselSection {
        ...CustomProductCarousel_UniqueFields
        productCarouselProducts: products(page: 1
        perPage: 16
        sort: { key: FEATURED, direction: DESC }) {
            ...CustomProductCarousel_ItemFields
            __typename
        }
        __typename
    }
    __typename
}

fragment AccordionSectionData on AccordionSection {
    __typename
    id
    title
    showTitle
    schema
    accordionBlocks {
        title
        text
        __typename
    }
}

fragment BreadcrumbSection on BreadcrumbSection {
    ...BreadcrumbDynamicSection
    __typename
}

fragment BreadcrumbDynamicSection on BreadcrumbSection {
    breadcrumbs {
        label
        url
        analyticsTitle
        __typename
    }
    __typename
}

fragment ListingBannerSection on ListingBannerSection {
    ...ListingBanner
    __typename
}

fragment ListingBanner on ListingBannerSection {
    layoutLegacy {
        fullWidth
        removePadding
        __typename
    }
    title
    description
    contrast
    logoImage
    backgroundImages {
        small {
            ...ImageAsset
            __typename
        }
        medium {
            ...ImageAsset
            __typename
        }
        large {
            ...ImageAsset
            __typename
        }
        __typename
    }
    __typename
}

fragment CategoryListingSection on CategoryListingSection {
    ...CategoryListing
    __typename
}

fragment CategoryListing on CategoryListingSection {
    title
    description
    thumbnailImage
    children {
        ...CategoryLeafSection
        __typename
    }
    hasCustomContent
    __typename
}

fragment CategoryLeafSection on CategoryListingChildren {
    title
    description
    thumbnailImage
    logoImage
    url
    ageRange
    tag
    thumbnailSrc {
        ...ImageAsset
        __typename
    }
    doesNotHaveAnAboutPage
    __typename
}

fragment CardContentSection on CardContentSection {
    ...CardContent
    __typename
}

fragment CardContent on CardContentSection {
    layoutLegacy {
        fullWidth
        removePadding
        __typename
    }
    moduleTitle
    showModuleTitle
    backgroundColor
    blocks {
        title
        isH1
        description
        textAlignment
        primaryLogoSrc {
            ...ImageAsset
            __typename
        }
        secondaryLogoSrc {
            ...ImageAsset
            __typename
        }
        logoPosition
        imageSrc {
            ...ImageAsset
            __typename
        }
        callToActionText
        callToActionLink
        callToActionUseAnalytics
        callToActionOpenInNewTab
        altText
        contrast
        videoMedia {
            ...VideoAssetFragment
            __typename
        }
        __typename
    }
    __typename
}

fragment VideoAssetFragment on VideoMedia {
    url
    id
    isLiveStream
    subtitlesUrl
    __typename
}

fragment CardCarouselSection on CardCarouselSection {
    ...CardCarouselContent
    __typename
}

fragment CardCarouselContent on CardCarouselSection {
    layoutLegacy {
        fullWidth
        removePadding
        __typename
    }
    moduleTitle
    showModuleTitle
    backgroundColor
    blocks {
        title
        isH1
        description
        textAlignment
        primaryLogoSrc {
            ...ImageAsset
            __typename
        }
        secondaryLogoSrc {
            ...ImageAsset
            __typename
        }
        logoPosition
        imageSrc {
            ...ImageAsset
            __typename
        }
        callToActionText
        callToActionLink
        callToActionUseAnalytics
        callToActionOpenInNewTab
        altText
        contrast
        videoMedia {
            ...VideoAssetFragment
            __typename
        }
        __typename
    }
    __typename
}

fragment CopyContent on CopyContentSection {
    layoutLegacy {
        fullWidth
        removePadding
        __typename
    }
    blocks {
        title
        body
        textAlignment
        titleColor
        imageSrc {
            ...ImageAsset
            __typename
        }
        __typename
    }
    __typename
}

fragment CopySectionData on CopySection {
    layoutLegacy {
        fullWidth
        removePadding
        __typename
    }
    title
    showTitle
    body
    __typename
}

fragment QuickLinksData on QuickLinkSection {
    layoutLegacy {
        fullWidth
        removePadding
        __typename
    }
    title
    quickLinks {
        title
        isH1
        link
        openInNewTab
        contrast
        imageSrc {
            ...ImageAsset
            __typename
        }
        __typename
    }
    __typename
}

fragment ContentBlockMixedData on ContentBlockMixed {
    layoutLegacy {
        fullWidth
        removePadding
        __typename
    }
    moduleTitle
    showModuleTitle
    blocks {
        title
        isH1
        description
        backgroundColor
        blockTheme
        contentPosition
        logoURL
        secondaryLogoURL
        logoPosition
        callToActionText
        callToActionLink
        altText
        backgroundImages {
            largeImage {
                small {
                    ...ImageAsset
                    __typename
                }
                large {
                    ...ImageAsset
                    __typename
                }
                __typename
            }
            smallImage {
                small {
                    ...ImageAsset
                    __typename
                }
                large {
                    ...ImageAsset
                    __typename
                }
                __typename
            }
            __typename
        }
        __typename
    }
    __typename
}

fragment UserGeneratedContentData on UserGeneratedContent {
    layoutLegacy {
        fullWidth
        removePadding
        __typename
    }
    ugcBlock {
        title
        text
        ugcType
        ugcKey
        __typename
    }
    __typename
}

fragment HeroBannerData on HeroBanner {
    layoutLegacy {
        fullWidth
        removePadding
        __typename
    }
    heroblocks {
        id
        title
        isH1
        tagline
        bannerTheme
        contentVerticalPosition
        contentHorizontalPosition
        contentHeight
        primaryLogoSrcNew {
            ...ImageAsset
            __typename
        }
        secondaryLogoSrcNew {
            ...ImageAsset
            __typename
        }
        videoMedia {
            ...VideoAssetFragment
            __typename
        }
        logoPosition
        contentBackground
        callToActionText
        callToActionLink
        callToActionOpenInNewTab
        brandedAppStoreAsset {
            ...ImageAsset
            __typename
        }
        callToActionUseAnalytics
        secondaryCallToActionText
        secondaryCallToActionLink
        secondaryBrandedAppStoreAsset {
            ...ImageAsset
            __typename
        }
        secondaryCallToActionUseAnalytics
        secondaryOpenInNewTab
        backgroundImagesNew {
            small {
                ...ImageAsset
                __typename
            }
            medium {
                ...ImageAsset
                __typename
            }
            large {
                ...ImageAsset
                __typename
            }
            __typename
        }
        altText
        showEmailSignupForm
        __typename
    }
    __typename
}

fragment GalleryData on Gallery {
    layoutLegacy {
        fullWidth
        removePadding
        __typename
    }
    galleryblocks {
        id
        contentHeight
        primaryLogoSrcNew {
            ...ImageAsset
            __typename
        }
        videoMedia {
            ...VideoAssetFragment
            __typename
        }
        backgroundImagesNew {
            small {
                ...ImageAsset
                __typename
            }
            medium {
                ...ImageAsset
                __typename
            }
            large {
                ...ImageAsset
                __typename
            }
            __typename
        }
        __typename
    }
    __typename
}

fragment MotionBannerData on MotionBanner {
    layoutLegacy {
        fullWidth
        removePadding
        __typename
    }
    motionBannerBlocks {
        id
        title
        isH1
        tagline
        bannerTheme
        contentHorizontalPosition
        primaryLogoSrc {
            ...ImageAsset
            __typename
        }
        secondaryLogoSrc {
            ...ImageAsset
            __typename
        }
        animatedMedia
        videoMedia {
            ...VideoAssetFragment
            __typename
        }
        logoPosition
        contentBackground
        callToActionText
        callToActionLink
        callToActionUseAnalytics
        secondaryCallToActionText
        secondaryCallToActionLink
        secondaryCallToActionUseAnalytics
        backgroundImages {
            small {
                ...ImageAsset
                __typename
            }
            medium {
                ...ImageAsset
                __typename
            }
            large {
                ...ImageAsset
                __typename
            }
            __typename
        }
        altText
        __typename
    }
    __typename
}

fragment MotionSidekickData on MotionSidekick {
    layoutLegacy {
        fullWidth
        removePadding
        __typename
    }
    motionSidekickBlocks {
        id
        title
        isH1
        tagline
        bannerTheme
        contentHorizontalPosition
        primaryLogoSrc {
            ...ImageAsset
            __typename
        }
        secondaryLogoSrc {
            ...ImageAsset
            __typename
        }
        animatedMedia
        videoMedia {
            ...VideoAssetFragment
            __typename
        }
        logoPosition
        contentBackground
        callToActionText
        callToActionLink
        callToActionUseAnalytics
        backgroundImages {
            small {
                ...ImageAsset
                __typename
            }
            medium {
                ...ImageAsset
                __typename
            }
            large {
                ...ImageAsset
                __typename
            }
            __typename
        }
        altText
        __typename
    }
    __typename
}

fragment InPageNavData on InPageNav {
    layoutLegacy {
        fullWidth
        removePadding
        __typename
    }
    inPageNavBlocks {
        id
        title
        isH1
        text
        contrast
        primaryLogoSrc
        secondaryLogoSrc
        animatedMedia
        videoMedia {
            url
            id
            subtitlesUrl
            __typename
        }
        contentBackground
        backgroundImages {
            small
            medium
            large
            __typename
        }
        callToActionText
        callToActionLink
        callToActionUseAnalytics
        openInNewTab
        secondaryCallToActionText
        secondaryCallToActionLink
        secondaryCallToActionUseAnalytics
        secondaryOpenInNewTab
        __typename
    }
    __typename
}

fragment TableData on TableSection {
    rows {
        isHeadingRow
        cells
        __typename
    }
    __typename
}

fragment RecommendationSectionData on RecommendationSection {
    __typename
    title
    showTitle
    recommendationType
}

fragment SidekickBannerData on SidekickBanner {
    __typename
    id
    layoutLegacy {
        fullWidth
        removePadding
        __typename
    }
    sidekickBlocks {
        title
        isH1
        text
        textAlignment
        contrast
        backgroundColor
        logoSrc {
            ...ImageAsset
            __typename
        }
        secondaryLogoSrc {
            ...ImageAsset
            __typename
        }
        logoPosition
        ctaTextPrimary: ctaText
        ctaLinkPrimary: ctaLink
        ctaOpenInNewTab
        ctaUseAnalyticsPrimary: ctaUseAnalytics
        brandedAppStoreAsset {
            ...ImageAsset
            __typename
        }
        ctaTextSecondary
        ctaLinkSecondary
        ctaOpenInNewTabSecondary
        ctaUseAnalyticsSecondary
        secondaryBrandedAppStoreAsset {
            ...ImageAsset
            __typename
        }
        contentHeight
        bgImages {
            large
            __typename
        }
        videoMedia {
            ...VideoAssetFragment
            __typename
        }
        altText
        __typename
    }
}

fragment ProductCarousel_UniqueFields on ProductCarouselSection {
    __typename
    productCarouselTitle: title
    showTitle
    showAddToBag
    seeAllLink
}

fragment TextBlockData on TextBlock {
    textBlocks {
        title
        isH1
        text
        textAlignment
        contrast
        backgroundColor
        callToActionLink
        callToActionText
        callToActionUseAnalytics
        openInNewTab
        secondaryCallToActionLink
        secondaryCallToActionText
        secondaryCallToActionUseAnalytics
        secondaryOpenInNewTab
        __typename
    }
    __typename
}

fragment TextBlockSEOData on TextBlockSEO {
    layoutLegacy {
        fullWidth
        removePadding
        __typename
    }
    textBlocks {
        title
        text
        __typename
    }
    __typename
}

fragment CrowdTwistWidgetSection on CrowdTwistWidgetSection {
    __typename
    id
    heading
    activityId
    rewardId
    defaultOpen
}

fragment CrowdTwistToggleWidgetSection on CrowdTwistToggleWidgetSection {
    __typename
    defaultOpen
    firstStepDescription
    firstStepHeading
    id
    secondStepHeading
    heading
    radioButtons {
        activityId
        buttonLabel
        rewardId
        __typename
    }
    isVipBannerVisible
}

fragment CrowdTwistCodeRedemptionBanner on CrowdTwistCodeRedemptionBanner {
    __typename
    heading
    description
    buttonText
    inputLabel
    campaignId
    groupCampaignId
    imageUrl
}

fragment CodedSection on CodedSection {
    __typename
    id
    componentName
    properties {
        key
        value
        __typename
    }
    text {
        key
        value
        __typename
    }
    media {
        key
        values {
            id
            contentType
            fileSize
            filename
            url
            title
            __typename
        }
        __typename
    }
}

fragment GridSectionData on GridSection {
    layout {
        colors {
            background
            __typename
        }
        __typename
    }
    layoutLegacy {
        fullWidth
        removePadding
        __typename
    }
    items {
        id
        image
        videoMedia {
            ...VideoAssetFragment
            __typename
        }
        href
        text
        textContrast
        __typename
    }
    __typename
}

fragment AudioSectionData on AudioSection {
    tracks {
        trackArt {
            ...ImageAsset
            __typename
        }
        src
        title
        description
        __typename
    }
    backgroundColor
    textContrast
    backgroundImage {
        mobile {
            ...ImageAsset
            __typename
        }
        desktop {
            ...ImageAsset
            __typename
        }
        __typename
    }
    seriesTitle
    seriesThumbnail {
        ...ImageAsset
        __typename
    }
    __typename
}

fragment StickyCTAData on StickyCTASection {
    layoutLegacy {
        fullWidth
        removePadding
        __typename
    }
    item {
        backgroundColor
        ctaBackgroundImage
        ctaPosition
        href
        openInNewTab
        text
        textAlign
        textContrast
        effect
        delay
        __typename
    }
    __typename
}

fragment MotionSidekick1x1Data on MotionSidekick1x1 {
    layoutLegacy {
        fullWidth
        removePadding
        __typename
    }
    motionSidekickBlocks {
        id
        title
        description
        textContrast
        contentHorizontalPosition
        primaryLogoSrc {
            ...ImageAsset
            __typename
        }
        secondaryLogoSrc {
            ...ImageAsset
            __typename
        }
        inlineVideo {
            ...VideoAssetFragment
            __typename
        }
        fullVideo {
            ...VideoAssetFragment
            __typename
        }
        logoHorizontalPosition
        backgroundColor
        primaryCallToActionText
        primaryCallToActionLink
        primaryCallToActionUseAnalytics
        secondaryCallToActionText
        secondaryCallToActionLink
        secondaryCallToActionUseAnalytics
        __typename
    }
    __typename
}

fragment ImageTransitionSliderData on ImageTransitionSlider {
    layoutLegacy {
        fullWidth
        removePadding
        __typename
    }
    imageTransitionSliderBlocks {
        id
        title
        description
        backgroundColor
        contrast
        ctas {
            link
            text
            useAnalytics
            __typename
        }
        contentHorizontalPosition
        firstImage {
            ...ImageAsset
            __typename
        }
        secondImage {
            ...ImageAsset
            __typename
        }
        __typename
    }
    __typename
}

fragment ImageXrayViewerData on ImageXrayViewer {
    layoutLegacy {
        fullWidth
        removePadding
        __typename
    }
    imageXrayViewerBlocks {
        id
        title
        description
        backgroundColor
        contrast
        ctas {
            link
            text
            useAnalytics
            __typename
        }
        contentHorizontalPosition
        firstImage {
            ...ImageAsset
            __typename
        }
        secondImage {
            ...ImageAsset
            __typename
        }
        __typename
    }
    __typename
}

fragment PollsSectionData on PollsSection {
    id
    question
    backgroundColor
    answerFillColor
    answerBorderColor
    answers {
        answer
        id
        __typename
    }
    image {
        ...ImageAsset
        __typename
    }
    imageAlignment
    pollResults {
        answers {
            answerId
            count
            __typename
        }
        totalVotes
        __typename
    }
    showPollResults
    submissionConfirmationTitle
    submissionConfirmationContent
    __typename
}

fragment ArtNavigationData on ArtNavigation {
    layoutLegacy {
        fullWidth
        removePadding
        __typename
    }
    artNavigationBlocks {
        id
        title
        cardTitle
        darkMode
        callToActionLink
        backgroundImage {
            ...ImageAsset
            __typename
        }
        logoImage {
            ...ImageAsset
            __typename
        }
        textImage {
            ...ImageAsset
            __typename
        }
        __typename
    }
    __typename
}

fragment MotionBanner16x9Data on MotionBanner16x9 {
    layoutLegacy {
        fullWidth
        removePadding
        __typename
    }
    motionBannerBlocks {
        id
        title
        isH1
        tagline
        contentHorizontalPosition
        primaryLogoSrc {
            ...ImageAsset
            __typename
        }
        secondaryLogoSrc {
            ...ImageAsset
            __typename
        }
        animatedMedia
        videoMedia {
            ...VideoAssetFragment
            __typename
        }
        logoPosition
        contentBackground
        callToActionText
        callToActionLink
        callToActionUseAnalytics
        secondaryCallToActionText
        secondaryCallToActionLink
        secondaryCallToActionUseAnalytics
        altText
        __typename
    }
    __typename
}

fragment QuickLinksAdvancedData on QuickLinkAdvancedSection {
    layoutLegacy {
        fullWidth
        removePadding
        __typename
    }
    linkCount
    backgroundColor
    items {
        title
        link
        openInNewTab
        contrast
        imageSrc {
            small {
                ...ImageAsset
                __typename
            }
            medium {
                ...ImageAsset
                __typename
            }
            __typename
        }
        textAlignment
        textAlignmentVertical
        __typename
    }
    __typename
}

fragment ArticleSectionData on ArticleSection {
    id
    layoutLegacy {
        fullWidth
        removePadding
        __typename
    }
    articleBlocks {
        id
        contentTitle
        setAsH1
        richText
        width
        product {
            ...Product_ProductItem
            __typename
        }
        productAlignment
        backgroundColor
        contentAlignment
        callToActionText
        callToActionType
        callToActionLink
        callToActionUseAnalytics
        openInNewTab
        image {
            ...ImageAsset
            __typename
        }
        caption
        captionDarkMode
        __typename
    }
    __typename
}

fragment RelatedArticleSectionData on RelatedArticleSection {
    id
    title
    articles {
        id
        title
        description
        url
        image {
            ...ImageAsset
            __typename
        }
        __typename
    }
    backgroundColor
    showCta
    target {
        text
        href
        __typename
    }
    __typename
}

fragment FeatureExplorerSectionData on FeatureExplorerSection {
    id
    title
    showHeader
    showHeaderLabel
    backgroundGradientColors {
        backgroundLightColor
        backgroundDarkColor
        __typename
    }
    overlayBackgroundColor
    overlayTextColor
    accentColor
    logo {
        image
        altText
        __typename
    }
    secondaryLogo {
        image
        altText
        __typename
    }
    features {
        title
        text
        scene
        position {
            x
            y
            __typename
        }
        video
        image
        __typename
    }
    frames
    __typename
}

fragment IdeaGeneratorSectionData on IdeaGeneratorSection {
    id
    title
    previewContent {
        title
        text
        callToActionText
        __typename
    }
    mainContent {
        startText
        retryText
        ideaLimit
        unlockThreshold
        endText
        callToAction {
            text
            link
            openInNewWindow
            __typename
        }
        __typename
    }
    problems {
        text
        image
        altText
        tags
        __typename
    }
    multipliers {
        text
        image
        altText
        tags
        __typename
    }
    validProblems {
        text
        image
        altText
        tags
        __typename
    }
    __typename
}

fragment TabbedContentExplorerData on TabbedContentExplorerSection {
    __typename
    id
    blocks {
        title
        backgroundColor
        accentColor
        target {
            href
            text
            openInNewTab
            __typename
        }
        images {
            alt
            desktop {
                ...ImageAsset
                __typename
            }
            mobile {
                ...ImageAsset
                __typename
            }
            __typename
        }
        __typename
    }
}

fragment CustomProductCarousel_UniqueFields on CustomProductCarouselSection {
    __typename
    productCarouselTitle: title
    showTitle
    showAddToBag
    seeAllLink
    backgroundColor
}

fragment CustomProductCarousel_ItemFields on CustomProductCarouselItem {
    product {
        ...Product_ProductItem
        __typename
    }
    imageOverride {
        ...ImageAsset
        __typename
    }
    imageBackgroundColor
    contentBackgroundColor
    ctaButtonColor
    __typename
}

fragment Countdown on CountdownBannerChild {
    title
    isH1
    text
    textPosition
    textAlignment
    contrast
    backgroundColor
    callToActionLink
    callToActionText
    openInNewTab
    countdownDate
    __typename
}

fragment CountdownBannerData on CountdownBanner {
    layoutLegacy {
        fullWidth
        removePadding
        __typename
    }
    countdownBannerBlocks {
        ...Countdown
        __typename
    }
    __typename
}

fragment CardContentRTWData on CardContentRTWSection {
    layoutLegacy {
        fullWidth
        removePadding
        __typename
    }
    moduleTitle
    showModuleTitle
    backgroundColor
    preferCarousel
    hasShadow
    blocks {
        title
        description
        backgroundColor
        textAlignment
        imageSrc {
            ...ImageAsset
            __typename
        }
        altText
        videoMedia {
            ...VideoAssetFragment
            __typename
        }
        __typename
    }
    __typename
}

fragment ExpandedCardContentData on ExpandedCardContentSection {
    isStaggered
    darkMode
    contentOpacity
    contentBackgroundColor
    blocks {
        id
        title
        text
        darkMode
        contentAlignment
        desktopBackgroundImage {
            ...ImageAsset
            __typename
        }
        mobileBackgroundImage {
            ...ImageAsset
            __typename
        }
        target {
            href
            text
            openInNewTab
            __typename
        }
        __typename
    }
    __typename
}

fragment ArticleTextData on ArticleTextSection {
    id
    title
    richText
    textAlignment
    __typename
}

fragment ArticleImageSectionData on ArticleImageSection {
    title
    image {
        ...ImageAsset
        __typename
    }
    altText
    caption
    captionBackground
    __typename
}

fragment ExpandedProductLeafData on ExpandedProductLeafSection {
    id
    productCode
    __typename
}

fragment Dots3DSectionData on Dots3DSection {
    id
    isSingleSKU
    tiles {
        id
        color
        shape
        sticker
        sets
        ip
        rotation
        alpha
        __typename
    }
    tiles3D
    models {
        setId
        ip
        label
        pdpLink
        glb
        image
        zoom {
            min
            max
            start
            __typename
        }
        cameraRotation {
            minHorizontal
            maxHorizontal
            minVertical
            maxVertical
            __typename
        }
        rotation
        bgColor
        bgContrast
        backLight
        __typename
    }
    home {
        title
        heading
        body
        selectButtonText
        backgroundImageMobile
        backgroundImageTablet
        backgroundImageDesktop
        __typename
    }
    decorate {
        title
        tileDrawer {
            title
            filterText
            colorPickerText
            __typename
        }
        helpGuide {
            title
            closeButtonText
            helpItems {
                title
                text
                image
                __typename
            }
            __typename
        }
        helpText
        clearTilesText
        __typename
    }
    download {
        title
        logo
        discoverText
        discoverImage
        discoverButtonText
        discoverButtonLink
        downloadButtonText
        ctaButtonText
        canvasLabel
        backgroundImageMobile
        backgroundImageTablet
        backgroundImageDesktop
        hashtag
        __typename
    }
    popups {
        name
        title
        text
        continueText
        cancelText
        __typename
    }
    logos {
        ip
        logo
        __typename
    }
    modal {
        ariaLabel
        closeButtonAriaLabel
        __typename
    }
    __typename
}

fragment NinetiethAnniversaryExperienceData on NinetiethAnniversaryExperienceSection {
    start {
        heading
        subHeading
        text
        ctaText
        backgroundImage
        __typename
    }
    charge {
        instructionText
        skipText
        ctaText
        chargeMask
        __typename
    }
    end {
        downloadButtonText
        restartButtonText
        straplineText
        hashtagText
        endScreenBackgroundTopLeft
        endScreenBackgroundBottomRight
        __typename
    }
    bricks {
        image
        name
        quote
        released
        dimensions
        count
        weight
        __typename
    }
    quiz {
        instructionText
        questionCount
        questions {
            text
            image
            swipeRightText
            swipeLeftText
            colorPalette {
                background
                question
                card
                topLeftImage
                bottomRightImage
                __typename
            }
            __typename
        }
        __typename
    }
    brickCard {
        countLabel
        backgroundImage
        colors
        dimensionsLabel
        heading
        releasedLabel
        weightLabel
        __typename
    }
    __typename
}

fragment ArticleGroupSectionData on ArticleGroupSection {
    articleGroupBlocks {
        ... on ArticleGroupImageBlock {
            id
            __typename
            layout {
                width
                colors {
                    background
                    __typename
                }
                __typename
            }
            image {
                ...ImageAsset
                __typename
            }
            altText
            caption
            captionBackground
        }
        ... on ArticleGroupTextBlock {
            id
            __typename
            layout {
                width
                colors {
                    background
                    __typename
                }
                __typename
            }
            textAlignment
            richText
        }
        __typename
    }
    __typename
}

fragment MotionBannerSectionData on MotionBannerSection {
    id
    title
    isH1
    tagline
    bannerTheme
    contentHorizontalPosition
    contentVerticalPosition
    aspectRatio
    backgroundVideoMobile
    backgroundVideoTablet
    backgroundVideoDesktop
    modalVideo {
        ...VideoAssetFragment
        __typename
    }
    backgroundImages {
        small {
            ...ImageAsset
            __typename
        }
        medium {
            ...ImageAsset
            __typename
        }
        large {
            ...ImageAsset
            __typename
        }
        __typename
    }
    altText
    primaryLogoSrc {
        ...ImageAsset
        __typename
    }
    secondaryLogoSrc {
        ...ImageAsset
        __typename
    }
    primaryCallToAction {
        href
        text
        openInNewTab
        __typename
    }
    secondaryCallToAction {
        href
        text
        openInNewTab
        __typename
    }
    logoPosition
    __typename
}

fragment ColorPaletteColor on ColorPaletteColor {
    designToken
    contrastingTextColor
    __typename
}

fragment AdvancedProductHeroBannerSectionData on AdvancedProductHeroBannerSection {
    id
    title
    productCode
    productDescriptionText
    isAtTopOfPage
    infoTagText
    target {
        href
        text
        __typename
    }
    colorPalette {
        primary {
            ...ColorPaletteColor
            __typename
        }
        secondary {
            ...ColorPaletteColor
            __typename
        }
        tertiary {
            ...ColorPaletteColor
            __typename
        }
        quaternary {
            ...ColorPaletteColor
            __typename
        }
        __typename
    }
    productImage {
        desktop
        mobile
        alt
        __typename
    }
    infoPanel {
        text
        alt
        logo
        secondaryAlt
        secondarylogo
        __typename
    }
    __typename
}

fragment PlayTypeDetectorSectionData on PlayTypeDetectorSection {
    id
    colors {
        answer1
        answer2
        cardTheme
        cta
        __typename
    }
    startStep {
        heading
        text
        ctaText
        background {
            desktopImage
            desktopTheme
            mobileImage
            mobileTheme
            __typename
        }
        __typename
    }
    questionsStep {
        background {
            desktopImage
            desktopTheme
            mobileImage
            mobileTheme
            __typename
        }
        slider {
            helpText
            draggerImage
            answer1TargetImage
            answer1ConfirmationImage
            answer2TargetImage
            answer2ConfirmationImage
            __typename
        }
        questions {
            id
            text
            answers {
                text
                image
                targetPage
                targetQuestion
                __typename
            }
            __typename
        }
        __typename
    }
    referralStep {
        image
        heading
        text
        __typename
    }
    __typename
}

fragment SocialShareSectionData on SocialShareSection {
    id
    title
    sharingTitle
    sharingBodyText
    socialSharePlatforms
    hashtags
    displaySocialIcons
    layout {
        width
        __typename
    }
    __typename
}

fragment EcosystemJourneyStarterData on EcosystemJourneyStarterSection {
    id
    title
    isH1
    icmp_tag
    themes {
        colors {
            background
            contentTitle
            contentBackground
            contentBodyMobile
            contentBodyDesktop
            navigationHighlight
            accentLeft
            accentRight
            cta
            navigation
            __typename
        }
        navigation {
            name
            logo
            thumbnail
            __typename
        }
        content {
            image
            altText
            title
            body
            ageSuitability
            primaryCTALabel
            primaryCTALink
            primaryCTAOpenInNewWindow
            secondaryCTALabel
            secondaryCTALink
            secondaryCTAOpenInNewWindow
            __typename
        }
        productCarousel {
            label
            products {
                productCode
                image
                openInNewWindow
                __typename
            }
            __typename
        }
        __typename
    }
    __typename
}

fragment ResultsBannerSectionData on ResultsBannerSection {
    id
    title
    resultsBannerBackgroundColor: backgroundColor
    fontColor
    mobileImageLeft
    desktopImageLeft
    desktopImageRight
    desktopButtonText
    __typename
}

fragment CharacterExplorerSectionData on CharacterExplorerSection {
    id
    title
    backButtonLabel
    productSubtitle
    characters {
        intro {
            image
            tagline
            characterCTALabel
            title
            __typename
        }
        colors {
            character
            characterLabel
            bio
            bioLabel
            background
            textHighlight
            highlightLabelDesktop
            highlightLabelMobile
            productBackground
            productLabel
            __typename
        }
        characterProduct {
            title
            overrideImage
            code
            __typename
        }
        content {
            body
            header
            bioImage
            mobileNavigationLabels
            passions {
                icon
                value
                __typename
            }
            __typename
        }
        nonShoppableContent {
            image
            text
            mobileNavigationLabelRight
            __typename
        }
        __typename
    }
    __typename
}

fragment SubmissionFormBannerSectionData on SubmissionFormBannerSection {
    id
    title
    layout {
        width
        colors {
            background
            __typename
        }
        __typename
    }
    bannerContent {
        title
        body
        backgroundImageSmall
        backgroundImageMedium
        backgroundImageLarge
        backgroundColorMobile
        fontColor
        video
        primaryCTALabel
        secondaryCTALabel
        __typename
    }
    modalContent {
        formId
        closeLabel
        title
        description
        fields {
            fieldId
            type
            label
            placeholderText
            errorText
            options {
                value
                label
                __typename
            }
            __typename
        }
        termsAndConditionsLabel
        termsAndConditions
        termsAndConditionsFile
        termsAndConditionsDownloadLabel
        consentLabel
        submitButtonLabel
        __typename
    }
    successMessageContent {
        title
        body
        additionalContent
        ctaLabel
        ctaIcon
        ctaLink
        __typename
    }
    translations {
        key
        value
        __typename
    }
    __typename
}

fragment RebuildTheWorld2023Data on RebuildTheWorld2023 {
    content {
        common {
            backButtonText
            heroImageAltText
            flareImage
            magnifyingGlassImage
            backgroundColor
            __typename
        }
        bannerStep {
            title
            body
            ctaText
            foregroundImageMobile
            foregroundImageDesktop
            __typename
        }
        playStep {
            ctaText
            nextButtonText
            previousButtonText
            instructionalText
            unlockText
            errorText
            backgroundImageMobile
            backgroundImageDesktop
            incorrectGuessImage
            incorrectGuessAltText
            codeGroups {
                colorName
                title
                body
                defaultCode {
                    image
                    altText
                    __typename
                }
                correctCode {
                    image
                    altText
                    __typename
                }
                incorrectCodes {
                    image
                    altText
                    __typename
                }
                __typename
            }
            __typename
        }
        heroRevealStep {
            loadingText
            preRevealText
            ctaText
            powers {
                colorName
                color
                accentColor
                backgroundImageMobile
                backgroundImageDesktop
                __typename
            }
            __typename
        }
        customiseStep {
            closeButtonText
            personaliseButtonText
            finishButtonText
            traySections {
                head {
                    label
                    active
                    inactive
                    options
                    __typename
                }
                hair {
                    label
                    active
                    inactive
                    options
                    __typename
                }
                beard {
                    label
                    active
                    inactive
                    options
                    __typename
                }
                torso {
                    label
                    active
                    inactive
                    options
                    __typename
                }
                legs {
                    label
                    active
                    inactive
                    options
                    __typename
                }
                crutches {
                    label
                    active
                    inactive
                    options
                    __typename
                }
                __typename
            }
            __typename
        }
        confirmStep {
            flipButtonText
            rearCardTitle
            rearCardBodyAdults
            rearCardBodyKids
            rearCardCTAText
            rearCardCTALink
            downloadCTAText
            shareCTAText
            restartCTAText
            hashtagText
            rearCardImage
            frontCardImages {
                colorName
                image
                __typename
            }
            __typename
        }
        minifigLayers {
            head {
                color
                colorLabel
                image
                __typename
            }
            hair {
                color
                colorLabel
                images
                __typename
            }
            beard {
                color
                colorLabel
                images
                __typename
            }
            face
            torso {
                colorName
                images
                __typename
            }
            legs {
                colorName
                images
                __typename
            }
            crutches
            __typename
        }
        __typename
    }
    __typename
}

fragment ContentContainerData_WithGiftFinderResultsSection on ContentContainer {
    ...BaseContentContainer
    section {
        ...BaseContentSection
        ... on GiftFinderResultsSection {
            id
            layout {
                width
                colors {
                    background
                    __typename
                }
                containerType
                removePadding
                backgroundColor
                __typename
            }
            productResults(giftFinderFilters: $filters) {
                ...ProductListings
                __typename
            }
            __typename
        }
        __typename
    }
    __typename
}`

	variables := map[string]any{
		"offset": 0,
		"sort": map[string]any{
			"key":       "FEATURED",
			"direction": "DESC",
		},
		"filters":         []string{},
		"searchSessionId": 1,
		"perPage":         18,
		"page":            1,
		"slug":            getSlug(options),
	}

	gql := map[string]any{
		"query":     query,
		"variables": variables,
	}

	client := &http.Client{}
	jsonValue, _ := json.Marshal(gql)

	req, err := http.NewRequest(
		"POST",
		"https://www.lego.com/api/graphql/ContentPageQuery",
		strings.NewReader(string(jsonValue)),
	)

	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-locale", "en-US")

	return client.Do(req)
}
