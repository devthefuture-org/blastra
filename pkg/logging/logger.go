package logging

import (
	"os"

	log "github.com/sirupsen/logrus"
)

const asciiArt = `                                                                                                    
                                                                                                    
                                                                                                    
                                                                                                    
                                                                                                    
                                                                                      .             
                                                                                                    
                                                                                                    
                                                                                                    
                .                                        .                                          
                                                                   .  .                             
                                                                .                .@@%               
       .                                 :                                     .@@@@                
                                      +@@@@@.                        =@.      @@@@                  
                                   +@@@@@@@@@@@.             @@@    @@@@    @@@@                    
                                .@@@@@@.   -@@@@@@=           :      #.   %@@@                      
                             :@@@@@@:         %@@@@@    #@%             *@@@                        
                          :@@@@@@+               %-   *@@@    #@@.    -@@@.                         
                       .@@@@@@%                     -@@@    :@@@    .@@@:                           
                    .@@@@@@@                      -@@@    :@@@.    @@@*   .                         
                 .@@@@@@@                       *@@@.   :@@@.    @@@%   .@@@:                       
               @@@@@@@.          @@    .      %@@@.   :@@@.    @@@@      @@.   -                    
            %@@@@@@             @@@#   .    @@@@.   :@@@.    @@@@            .@@@                   
           @@@@@                          @@@@.   =@@@-    #@@@    =@@.       @@                    
           @@@-                         @@@@.   #@@@-    +@@@    .@@@*   .@                         
           @@@+                       @@@@.   @@@@.    -@@@     @@@@    @@@@          .             
           @@@%                     @@@@   .@@@@.    -@@@     @@@@    @@@@                          
           @@@%                   @@@@   :@@@@     *@@@    .@@@@    %@@@                            
           @@@@                .@@@%   *@@@#     %@@#     @@@@.   *@@@.                             
           @@@@              .@@@=   @@@@:     @@@%     @@@@.   +@@@.                               
           @@@@             @@@.  @@@@@     #@@@%     @@@@:   -@@@:                                 
           @@@@           @@@= @@@@@-    .@@@@@     @@@@-   :@@@+    @                              
           @@@@        -@@@@%@@@@*     @@@@@@     @@@@-   .@@@#    @@@                              
           @@@@    . +@@@@:        .@@@@@@@-    @@@@*   .@@@@     =@@@                .             
           @@@@     @@@%     :@@@@@@@@@@@@    #@@@%   .@@@%       -@@@                              
           @@@*    @@@.  .@@@@@@@@@@@@@@.   :@@@@    @@@@         -@@@                              
           @@@.   @@@.  @@@@@@@@@@@@@@@    @@@@.   @@@@      .    -@@@                              
           @@@   +@@@  @@@@@@@@@@@@@@-   @@@@:   @@@@             -@@@                              
           @@@   @@@.  @@@@@@@@@@@@@.  *@@@*   %@@@               :@@@                              
          .@@@   @@@=  @@@@@@@@@@@@.  +@@@   .@@@                 :@@@                              
           @@@%  .@@@  .@@@@@@@@@@   =@@%   @@@                  *@@@@            .                 
             @@   @@@@   @@@@@@@.   @@@#   @@    ..           %@@@@@@.                              
                   @@@@.          @@@@-         %@@        #@@@@@@-                                 
                    @@@@@@@+-=%@@@@@@                   +@@@@@@-                                    
                      %@@@@@@@@@@@.                  +@@@@@@-                                       
                            .                     +@@@@@@=              .                           
                                .@@@         . -@@@@@@:         .                                   
                                %@@@@@@     .@@@@@@:                                                
                             .     @@@@@@@@@@@@@:                                                   
                                      @@@@@@@-                                                      
                                        .@:                                                         
        .                                                                                           
                                                .                                                   
                                                                                                    
                                                                         .                             


                        █▄▄ █░░ ▄▀█ █▀ ▀█▀ █▀█ ▄▀█
                        █▄█ █▄▄ █▀█ ▄█ ░█░ █▀▄ █▀█

    █▀▀ ▄▀█ █▀ ▀█▀ ░   █▀█ █▀▀ █░░ █ ▄▀█ █▄▄ █░░ █▀▀ ░   █▀ ▀█▀ █▀▀ █░░ █░░ ▄▀█ █▀█
    █▀░ █▀█ ▄█ ░█░ ▄   █▀▄ ██▄ █▄▄ █ █▀█ █▄█ █▄▄ ██▄ ▄   ▄█ ░█░ ██▄ █▄▄ █▄▄ █▀█ █▀▄`

var isTextFormatter bool

func ConfigureLogging() {
	logLevel := os.Getenv("BLASTRA_LOG_LEVEL")
	if logLevel == "" {
		logLevel = os.Getenv("LOG_LEVEL")
	}
	if logLevel == "" {
		logLevel = "info"
	}
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Warnf("Invalid log level %s, defaulting to info", logLevel)
		level = log.InfoLevel
	}
	log.SetLevel(level)

	// Directly output logs for immediate visibility:
	log.SetOutput(os.Stdout)

	// Check if we're running in a terminal
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		// Terminal detected - use text formatter with colors
		log.SetFormatter(&log.TextFormatter{
			ForceColors:   true,
			FullTimestamp: true,
		})
		isTextFormatter = true
	} else {
		// No terminal - use JSON formatter
		log.SetFormatter(&log.JSONFormatter{})
		isTextFormatter = false
	}

	log.Infof("Starting Blastra server with log level: %s", level)
}

func LogAsciiArt() {
	if !isTextFormatter {
		return
	}

	// Print ASCII art directly to stdout to avoid log formatting
	os.Stdout.Write([]byte("\n"))
	os.Stdout.Write([]byte(asciiArt))
	os.Stdout.Write([]byte("\n"))
}
