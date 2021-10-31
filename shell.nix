with import <nixpkgs> {}; 

mkShell {
  nativeBuildInputs = [ 
    buildPackages.git
    buildPackages.go_1_17
  ];
}
