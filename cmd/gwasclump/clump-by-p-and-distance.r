list.of.packages <- c("data.table", "optparse")
new.packages <- list.of.packages[!(list.of.packages %in% installed.packages()[,"Package"])]
if(length(new.packages)) install.packages(new.packages,repos = "http://cran.us.r-project.org")

library("optparse")
library("data.table")

option_list = list(
    make_option(c("-f", "--file"), type="character", default=NULL, help="BOLT output file"),
    make_option(c("-p", "--pvalue"), type="numeric", default=5e-8, help="P value threshold, below which loci are considered significant"),
    make_option(c("-r", "--locusRadius"), type="numeric", default=NULL, help="Number of bases around each lead SNP to exclude other potential lead SNPs")
)

opt <- parse_args(OptionParser(option_list=option_list))


message("Options enabled:")
message(paste(opt, collapse="\n"))

snpDistanceCutoff <- function(pos, bases) {
    lower = pos-bases
    if(lower<1){
        lower = 1
    }
    
    upper = pos+bases
    
    list(lower = lower, upper = upper)
}

pruneSNPs <- function(dat, flanking = 5e5, significance = 5e-8) {
    # Exit early if none are left in the input
    if(nrow(dat)==0){
        return()
    }
    
    # The best remaining SNP
    best <- subset(dat, is.finite(CHISQ_BOLT_LMM) & (CHISQ_BOLT_LMM == max(dat$CHISQ_BOLT_LMM)))
    if(nrow(best)==0){
        return()
    }
    
    # If the best remaining SNP is outside of your permitted bounds, exit
    if(best$P_BOLT_LMM > significance) {
        return()
    }
    
    # Identify the window around the SNP
    dist = snpDistanceCutoff(best$BP, flanking)

    # Destroy the neighboring SNPs
    remaining <- subset(dat, CHR != best$CHR | BP < dist$lower | BP > dist$upper )

    snpsDestroyed = nrow(dat) - nrow(remaining)

    # Track that:
    best$Neighbors = snpsDestroyed
    
    # Return the SNP and the remaining data with the 
    # flanking SNPs around the newest best SNP removed
    list(best = best, remaining = remaining)
}

iteratively.prune <- function(df, flanking=1e6, significance=1e-7) {
    bestSNPs <- list()

    testDat <- df
    i <- 0
    repeat {
        i <- i + 1
        res <- pruneSNPs(testDat, flanking, significance)
        if(is.null(res)) {
            break
        }

        bestSNPs[[i]] <- res$best
        testDat <- res$remaining
    }

    all.best <- do.call(rbind, bestSNPs)
    if(is.data.frame(all.best) && nrow(all.best) > 0) {
        all.best <- all.best[with(all.best, order(CHR, BP)),]
    }
    all.best
}

dat <- fread(opt$file, showProgress=TRUE)

pruned <- iteratively.prune(dat, significance = opt$pvalue, flanking=opt$locusRadius)
N_Loci = nrow(pruned)

if(is.null(N_Loci)) {
    N_Loci = 0
} else if(!is.finite(N_Loci)) {
    N_Loci = 0
}

res <- data.frame(
    File = opt$file,
    Cutoff = opt$pvalue,
    N_Loci = N_Loci
)

message(paste(res, collapse="\t"))

write.table(pruned, stdout(), sep = "\t", row.names = FALSE, quote = FALSE)

#pruned
