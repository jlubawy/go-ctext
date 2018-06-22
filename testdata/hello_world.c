
#include <stdio.h>
/*******************************************************************************
 * hello_world.c
 */ #include <stdio.h>

/* PUTS
 * multi-line comment
 */
#define PUTS( _s ) { \
    fputs( _s, stdout ); \
}

// Main function
int
main( void )
{
    PUTS( "Hello World 1\n" ); // comment 1
    PUTS( "Hello World 2\n" ); // comment 2
    PUTS(
        "Hello // World 2\n"
    ); // prints "Hello // World 2 \n"
    return 0;
}

#if 0
/* Allow nested /* comments even though not supported by most compilers */ */
#endif
