#include "go_asm.h"
#include "textflag.h"
                                  
DATA WCTABLE<>+0x000(SB)/8, $0x0000000000000000
DATA WCTABLE<>+0x008(SB)/8, $0x00FFFF0000FF0000  // WRONG tab, new line, carriage return
DATA WCTABLE<>+0x010(SB)/8, $0x0000000000000000
DATA WCTABLE<>+0x018(SB)/8, $0x00FFFF0000FF0000  // WRONG tab, new line, carriage return
GLOBL WCTABLE<>(SB), RODATA, $0x32

DATA SPTABLE<>+0x000(SB)/8, $0x0000000000FF0000  // space
DATA SPTABLE<>+0x008(SB)/8, $0x0000000000000000
DATA SPTABLE<>+0x010(SB)/8, $0x0000000000FF0000  // space
DATA SPTABLE<>+0x018(SB)/8, $0x0000000000000000
GLOBL SPTABLE<>(SB), RODATA, $0x32

TEXT ·IndexNotWC(SB), NOSPLIT, $0-72
    // first slice is 0-23 byte
	MOVQ a_base+0(FP), SI
	MOVQ a_len+8(FP), AX
	LEAQ ret_b+72(FP), R9

    // second slice is 24-47 byte
	//MOVQ b_base+24(FP), DI
	//MOVQ b_len+32(FP), BX
	//VMOVDQU (DI), Y1 // Y1 – destination
	
    // slice and 
	MOVQ ret_base+48(FP), R8

	VMOVDQU SPTABLE<>(SB), Y1
	//VMOVDQU ANDTABLE<>(SB), Y2

	
	//MOVQ $0x0F0F0F0F0F0F0F0F, DX
	MOVQ $0xF0F0F0F0F0F0F0F0, DX
	MOVQ DX, X0
	VPBROADCASTB X0, Y2 // Y2 – AND

	MOVQ $0xFFFFFFFFFFFFFFFF, DX
	MOVQ DX, X0
	VPBROADCASTB X0, Y5 //

    MOVQ SI, DI
	LEAQ -32(SI)(AX*1), R11
	
loop:   
	VMOVDQU (DI), Y3    // load string data
	VPAND Y3, Y2, Y3    // cleanup 4 bits
	VPSRLQ $4, Y3, Y3   // shift
	VPSHUFB Y3, Y1, Y3	// make a lookup
	VPANDN Y5, Y3, Y3
	//VMOVDQU Y3, (R8)
	VPTEST Y3, Y3       // test if result isn't zero	
    JNZ success
	
	ADDQ $32, DI
	CMPQ DI, R11
	JLT loop
	
	MOVQ R11, DI
	VMOVDQU (DI), Y3    // load string data
	VPAND Y3, Y2, Y3    // cleanup 4 bits
	VPSRLQ $4, Y3, Y3   // shift
	VPSHUFB Y3, Y1, Y3	// make a lookup
	VPANDN Y5, Y3, Y3
	VPTEST Y3, Y3       // test if result isn't zero	
    JNZ success
    
	MOVQ $(-1), (R9)
	VZEROUPPER
	RET

success:
	VPMOVMSKB Y3, DX
	BSFL DX, DX
	SUBQ SI, DI
    ADDQ DI, DX
    //VMOVDQU Y3, (R8)
    MOVQ DX, (R9)
	VZEROUPPER
	RET

