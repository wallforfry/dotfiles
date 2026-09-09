telemetry_failure_message() {
  case $1 in
    2) printf '%s\n' 'dérive de format des transcripts : activation et adhérence non mesurées' ;;
    3) printf '%s\n' 'lecture des transcripts interrompue : activation et adhérence non mesurées' ;;
    *) printf '%s\n' 'échec de la télémétrie : activation et adhérence non mesurées' ;;
  esac
}
