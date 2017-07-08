
// This is a fake mex.h just to let clang compile MEX cc files

#ifndef MEX_H
#define MEX_H

typedef struct mxArray_tag mxArray;
typedef size_t mwSize;

typedef enum {mxREAL, mxCOMPLEX} mxComplexity;

extern void mexErrMsgIdAndTxt(const char*, const char*, ...);

extern size_t mxGetM(const mxArray*);
extern size_t mxGetN(const mxArray*);

extern bool mxIsDouble (const mxArray*);
extern bool mxIsComplex(const mxArray*);

extern double* mxGetPr(const mxArray*);

extern mxArray* mxCreateDoubleMatrix(size_t, size_t, mxComplexity);

#endif // MEX_H

