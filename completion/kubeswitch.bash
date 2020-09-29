_kubeswitchcomplete()
{
  local first second

  first=${COMP_WORDS[COMP_CWORD]}
  second=${COMP_WORDS[COMP_CWORD-1]}

  case ${COMP_CWORD} in
    1)
      cmds="context ctx namespace ns"
      COMPREPLY=($(printf "%s\n" $cmds | grep -e "^$first"))
      ;;
    2)
      case ${second} in
        context)
          COMPREPLY=($(kubeswitch ctx -P | grep -e "^$first"))
          ;;
        ctx)
          COMPREPLY=($(kubeswitch ctx -P | grep -e "^$first"))
          ;;
        namespace)
          COMPREPLY=($(kubeswitch ns -P | grep -e "^$first"))
          ;;
        ns)
          COMPREPLY=($(kubeswitch ns -P | grep -e "^$first"))
          ;;
      esac
      ;;
    *)
      COMPREPLY=()
      ;;
  esac
}

complete -F _kubeswitchcomplete kubeswitch ks

