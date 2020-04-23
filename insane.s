#include "go_asm.h"
#include "textflag.h"

TEXT ·IndexNotWC(SB), NOSPLIT, $0-80
    // first slice is 0-23 byte
	MOVQ a_base+0(FP), SI
	MOVQ a_len+8(FP), AX

    // second slice is 24-47 byte
	MOVQ b_base+24(FP), DI
	MOVQ b_len+32(FP), BX
	VMOVDQU (DI), Y1 // Y1 – destination

    // slice and 
	MOVQ ret_base+48(FP), R8
	LEAQ ret_b+72(FP), R9
	
	MOVQ $0x0F0F0F0F0F0F0F0F, DX
	MOVQ DX, X0
	VPBROADCASTB X0, Y2 // Y2 – AND

    MOVQ SI, DI
	LEAQ -32(SI)(AX*1), R11
	
loop:   
	VMOVDQU (DI), Y3    // load string data
	VPAND Y3, Y2, Y3    // cleanup higher 4 bits
    VMOVDQA Y1, Y4      // copy mask
	VPSHUFB Y3, Y4, Y3	// make a lookup
	VPTEST Y3, Y3       // test if result isn't zero	
    JNZ success
	
	ADDQ $32, DI
	CMPQ DI, R11
	JLT loop
	
	VMOVDQU (DI), Y3    // load string data
	VPAND Y3, Y2, Y3    // cleanup higher 4 bits
    VMOVDQA Y1, Y4      // copy mask
	VPSHUFB Y3, Y4, Y3	// make a lookup
	VPTEST Y3, Y3       // test if result isn't zero	
    JNZ success
    
	VMOVDQU Y5, (R8)
	MOVQ $(-1), (R9)
	VZEROUPPER
	RET

success:
	VPMOVMSKB Y3, DX
	BSFL DX, DX
	SUBQ SI, DI
    ADDQ DI, DX
    VMOVDQU Y3, (R8)
    MOVQ DX, (R9)
	VZEROUPPER
	RET

