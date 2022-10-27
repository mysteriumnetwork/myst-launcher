set PATH=%PATH%;C:\Program Files (x86)\WiX Toolset v3.11\bin
set PATH=%PATH%;C:\Program Files (x86)\NirSoft\IconsExtract


del installer\*.msi installer\*.wixobj installer\*.ico
iconsext.exe /save "bin\myst-launcher-amd64.exe" "installer\" -icons


candle installer\installer.wxs installer\licenseDialogue.wxs  -arch x64 -out installer\
light installer\installer.wixobj installer\licenseDialogue.wixobj -dcl:high -ext WixUIExtension.dll -ext WixUtilExtension.dll -out installer\launcher-x64.msi
