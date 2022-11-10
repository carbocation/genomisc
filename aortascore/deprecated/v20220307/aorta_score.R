# This software is released under the following BSD 3-Clause License:

# Copyright (c) 2018 - 2021 James Pirruccello and The General Hospital
# Corporation.  All rights reserved.

# Redistribution and use in source and binary forms, with or without
# modification, are permitted provided that the following conditions are met:

# 1. Redistributions of source code must retain the above copyright notice, this
#    list of conditions and the following disclaimer.

# 2. Redistributions in binary form must reproduce the above copyright notice,
#    this list of conditions and the following disclaimer in the documentation
#    and/or other materials provided with the distribution.

# 3. Neither the name of the copyright holder nor the names of its contributors
#    may be used to endorse or promote products derived from this software
#    without specific prior written permission.

# THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
# AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
# IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
# DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
# FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
# DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
# SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
# CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
# OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
# OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

# AORTA Score. This file represents the R program for computing the AORTA Score
# v20220208. The script contains the model coefficients that were derived by
# James Pirruccello against UK Biobank aortic measurements using the glinternet
# package, from Lim and Hastie 2015.

# The main function is aorta.score.one, which expects a single-person's data in a
# single row from a dataframe. E.g., to score the 6th person in a dataframe:
#
# aorta.score.one(dat[6,])

# To apply the function to a dataframe of people, and save the predicted aortic
# diameter as a column:
#
# dat$predicted_ascending_aorta_diam <- apply(dat, 1, aorta.score.one)

# The variables in your dataframe need to have the same names as those that were
# used in the model. To dynamically learn those column names, use the following
# incantation:
#
# cat("Required variables:\n\n")
# cat(paste0(required.columns(), collapse = "\n"))

categorical.categorical <- read.table(text = "level1	level2	value	namedCatEff1	namedCatEff2
0	0	0.0018642410985508	sex_Female	prevalent_dm
1	0	-0.0018642410985508	sex_Female	prevalent_dm
0	1	-0.0018642410985508	sex_Female	prevalent_dm
1	1	0.0018642410985508	sex_Female	prevalent_dm
0	0	-0.0135519810235069	prevalent_dm	prevalent_htn
1	0	0.0135519810235069	prevalent_dm	prevalent_htn
0	1	0.0135519810235069	prevalent_dm	prevalent_htn
1	1	-0.0135519810235069	prevalent_dm	prevalent_htn
0	0	0.002141243122696	prevalent_dm	prevalent_hld
1	0	-0.00214124312269599	prevalent_dm	prevalent_hld
0	1	-0.00214124312269599	prevalent_dm	prevalent_hld
1	1	0.002141243122696	prevalent_dm	prevalent_hld
", header = TRUE, stringsAsFactors = FALSE)

categorical.continuous <- read.table(text = "value	level	namedCatEff	namedContEff
0.000709351250579631	0	sex_Female	age
-0.000709351250579631	1	sex_Female	age
0.00041548114475475	0	prevalent_htn	age
-0.00041548114475475	1	prevalent_htn	age
0.00426229304181247	0	sex_Female	bmi
-0.00426229304181247	1	sex_Female	bmi
0.000461004228908038	0	prevalent_htn	bmi
-0.000461004228908038	1	prevalent_htn	bmi
-3.43665258951226e-05	0	prevalent_htn	pulse_rate
3.43665258951226e-05	1	prevalent_htn	pulse_rate
-2.55119304327748e-05	0	prevalent_hld	pulse_rate
2.55119304327748e-05	1	prevalent_hld	pulse_rate
0.000193132800481548	0	prevalent_htn	sbp
-0.000193132800481548	1	prevalent_htn	sbp
2.78948263423092e-05	0	prevalent_hld	sbp
-2.78948263423092e-05	1	prevalent_hld	sbp
-0.000253813188811231	0	sex_Female	dbp
0.000253813188811231	1	sex_Female	dbp
-0.000130570574632789	0	sex_Female	height_cm
0.000130570574632788	1	sex_Female	height_cm
4.41324573371244e-05	0	prevalent_htn	weight_kg
-4.41324573371244e-05	1	prevalent_htn	weight_kg
", header = TRUE, stringsAsFactors = FALSE)

categorical.main <- read.table(text = "value	level	namedCatEff
-0.208236516644537	0	sex_Female
-0.0500052474750719	1	sex_Female
0.0117257396538741	0	prevalent_dm
-0.0117257396538742	1	prevalent_dm
-0.0952430280091533	0	prevalent_htn
0.0854637589117299	1	prevalent_htn
0.00543199048665385	0	prevalent_hld
0.00396052862467056	1	prevalent_hld
", header = TRUE, stringsAsFactors = FALSE)

continuous.continuous <- read.table(text = "value	namedContEff1	namedContEff2
-6.74789970573941e-05	age	sbp
3.59345378203175e-05	age	dbp
2.04636140829908e-05	bmi	pulse_rate
-1.26986174383947e-05	bmi	sbp
-5.5122536886366e-05	bmi	weight_kg
-2.16393471628469e-05	pulse_rate	sbp
-2.8496808714647e-07	pulse_rate	height_cm
-8.05890197382759e-07	sbp	dbp
-5.58194977712159e-06	sbp	weight_kg
-4.47095146625718e-06	dbp	weight_kg
", header = TRUE, stringsAsFactors = FALSE)

continuous.main <- read.table(text = "value	effect	namedContEff
0.0169282883420514	1	age
-0.000543461244055947	3	pulse_rate
0.00494314498031087	4	sbp
0.00632388770802047	5	dbp
0.00691588718356602	6	height_cm
0.0061148395603283	7	weight_kg
0.0069043003257741	2	bmi
", header = TRUE, stringsAsFactors = FALSE)

intercept <- read.table(text = "Intercept
0.107164988292691
", header = TRUE, stringsAsFactors = FALSE)

required.columns <- function() {
    unique(c(
        continuous.main$namedContEff,
        continuous.continuous$namedContEff1,
        continuous.continuous$namedContEff2,
        categorical.main$namedCatEff,
        categorical.continuous$namedCatEff,
        categorical.continuous$namedContEff,
        categorical.categorical$namedCatEff1,
        categorical.categorical$namedCatEff2
    ))
}

# oneperson is a dataframe consisting of one row containing the necessary data
# fields
aorta.score.one <- function(oneperson){
    
    # Make sure that the data frame has all columns present
    have.names <- names(oneperson)
    missing.names <- required.columns()[!(required.columns() %in% have.names)]
    if(length(missing.names) > 0){
        warning("Missing columns: {", paste0(missing.names, collapse = ", "), "}")
        stop(".")
    }
    
    if(!is.data.frame(oneperson)) {
        # If called with 'apply', we get a transposed named list
        oneperson <- data.frame(t(oneperson))
    }

    # Get rid of unused columns
    oneperson <- oneperson[,required.columns()]
    
    # Orient to be 'tall' instead of wide
    t.oneperson <- cbind(row.names(t(oneperson)), t(oneperson))
    row.names(t.oneperson) <- NULL
    colnames(t.oneperson) <- c('field', 'user_value')
    t.oneperson <- as.data.frame(t.oneperson, stringsAsFactors = FALSE)
    t.oneperson <- subset(t.oneperson, field %in% required.columns()) # Dump unused columns
    t.oneperson$user_value[t.oneperson$user_value == FALSE] <- 0 # Set true/false to numeric values
    t.oneperson$user_value[t.oneperson$user_value == TRUE] <- 1 # Set true/false to numeric values
    t.oneperson$user_value <- as.numeric(t.oneperson$user_value) # Coerce all remaining values to numerics. Caution: will cause any remaining strings to become NA
    
    # Sum
    t.score <- 0 + intercept$Intercept

    # Cat
    x <- merge(
        t.oneperson, 
        categorical.main, 
        by.x = c('field', 'user_value'), 
        by.y = c('namedCatEff', 'level'))
    # x
    t.score <- t.score + sum(x$value)

    # Cont
    x <- merge(
        t.oneperson, 
        continuous.main, 
        by.x = c('field'), 
        by.y = c('namedContEff'))
    # x
    t.score <- t.score + sum(x$user_value * x$value)

    # CatCat
    x <- merge(
        categorical.categorical, 
        t.oneperson, 
        by.x = c('namedCatEff1', 'level1'),
        by.y = c('field', 'user_value'))
    x1 <- merge(
        x, 
        t.oneperson, 
        by.x = c('namedCatEff2', 'level2'),
        by.y = c('field', 'user_value'))
    # x1
    t.score <- t.score + sum(x1$value)

    # ContCont
    x <- merge(
        continuous.continuous,
        t.oneperson,
        by.x = c('namedContEff1'),
        by.y = c('field'),
        suffixes = c('', '_1')
    )
    x1 <- merge(
        x,
        t.oneperson,
        by.x = c('namedContEff2'),
        by.y = c('field'),
        suffixes = c('', '_2')
    )
    # x1
    t.score <- t.score + sum(x1$value * x1$user_value * x1$user_value_2)

    # CatCont
    x <- merge(
        categorical.continuous, 
        t.oneperson, 
        by.x = c('namedCatEff', 'level'),
        by.y = c('field', 'user_value')
    )
    # x
    x1 <- merge(
        x, 
        t.oneperson, 
        by.x = c('namedContEff'),
        by.y = c('field'),
        suffixes = c('', '_1')
    )
    # x1
    t.score <- t.score + sum(x1$user_value * x1$value)

    t.score
}
